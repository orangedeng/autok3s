package aws

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/orangedeng/autok3s/pkg/cluster"
	"github.com/orangedeng/autok3s/pkg/common"
	"github.com/orangedeng/autok3s/pkg/providers"
	putil "github.com/orangedeng/autok3s/pkg/providers/utils"
	"github.com/orangedeng/autok3s/pkg/types"
	typesaws "github.com/orangedeng/autok3s/pkg/types/aws"
	"github.com/orangedeng/autok3s/pkg/utils"

	"github.com/Masterminds/semver"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"
)

const (
	providerName = "aws"

	defaultUser              = "ubuntu"
	ipRange                  = "0.0.0.0/0"
	defaultSecurityGroupName = "autok3s"
	defaultDeviceName        = "/dev/sda1"
)

const (
	deployCCMCommand = "echo \"%s\" | base64 -d | tee \"%s/cloud-controller-manager.yaml\""
)

var (
	// The aws CCM requires specific version for each k8s version.
	// This is from https://github.com/kubernetes/cloud-provider-aws/tree/master#compatibility-with-kubernetes, so we can update
	// the version according to this docs.
	// The map key should be semver constraint and the value is the aws ccm version
	ccmVersionMap = map[string]string{
		">= 1.30": "v1.30.1",
		"~1.29":   "v1.29.3",
		"~1.28":   "v1.28.1",
		"~1.27":   "v1.27.1",
		"~1.26":   "v1.26.1",
		"~1.25":   "v1.25.3",
		"~1.24":   "v1.24.4",
		"~1.23":   "v1.23.6",
		"~1.22":   "v1.22.7",
		"~1.21":   "v1.21.6",
		"~1.20":   "v1.20.4",
		"~1.19":   "v1.19.0-alpha.1",
		"~1.18":   "v1.18.0-alpha.1",
		"< 1.18":  "v1.18.0-alpha.1",
	}
	// extra args for ccm with the k8s version
	ccmExtraArgs = map[string][]string{
		">= 1.21": {
			"--use-service-account-credentials=true",
			"--configure-cloud-routes=false",
		},
	}
	ccmTemplate = template.Must(template.New("aws-ccm").Parse(amazonCCMTmpl))
)

// Amazon provider amazon struct.
type Amazon struct {
	*cluster.ProviderBase `json:",inline"`
	typesaws.Options      `json:",inline"`
	client                *ec2.Client
}

func init() {
	providers.RegisterProvider(providerName, func() (providers.Provider, error) {
		return newProvider(), nil
	})
}

func newProvider() *Amazon {
	base := cluster.NewBaseProvider()
	base.Provider = providerName
	amazonProvider := &Amazon{
		ProviderBase: base,
	}
	if opt, ok := common.DefaultTemplates[providerName]; ok {
		amazonProvider.Options = opt.(typesaws.Options)
	}
	return amazonProvider
}

// GetProviderName returns provider name.
func (p *Amazon) GetProviderName() string {
	return p.Provider
}

// GenerateClusterName generates and returns cluster name.
func (p *Amazon) GenerateClusterName() string {
	p.ContextName = fmt.Sprintf("%s.%s.%s", p.Name, p.Region, p.GetProviderName())
	return p.ContextName
}

// GenerateManifest generates manifest deploy command.
func (p *Amazon) GenerateManifest() []string {
	if p.CloudControllerManager {
		return []string{fmt.Sprintf(deployCCMCommand,
			base64.StdEncoding.EncodeToString([]byte(getAWSCCMManifest(p.K3sVersion, p.ClusterCidr))), common.K3sManifestsDir)}
	}
	return nil
}

// CreateK3sCluster create K3S cluster.
func (p *Amazon) CreateK3sCluster() (err error) {
	if p.SSHUser == "" {
		p.SSHUser = defaultUser
	}
	return p.InitCluster(p.Options, p.GenerateManifest, p.generateInstance, nil, p.rollbackInstance, p.GenerateMasterExtraArgs, p.GenerateWorkerExtraArgs)
}

// UpgradeK3sCluster AWS Customized k3s upgrade
func (p *Amazon) UpgradeK3sCluster(clusterName, installScript, channel, version, packageName, packagePath string) error {
	if err := p.ProviderBase.UpgradeK3sCluster(clusterName, installScript, channel, version, packageName, packagePath); err != nil {
		return err
	}
	state, err := common.DefaultDB.GetCluster(clusterName, p.Provider)
	if err != nil {
		return err
	}
	// following error should never happen
	if state == nil {
		return fmt.Errorf("cluster %s is not exist", clusterName)
	}

	p.SetMetadata(&state.Metadata)
	_ = p.SetOptions(state.Options)

	if p.CloudControllerManager {
		c := common.ConvertToCluster(state, true)
		// using the new version to generate manifest.
		p.K3sVersion = version
		return p.DeployExtraManifest(&c, p.GenerateManifest())
	}
	return nil
}

// JoinK3sNode join K3S node.
func (p *Amazon) JoinK3sNode() (err error) {
	if p.SSHUser == "" {
		p.SSHUser = defaultUser
	}
	return p.JoinNodes(p.generateInstance, p.syncInstances, false, p.rollbackInstance, p.GenerateMasterExtraArgs, p.GenerateWorkerExtraArgs, nil)
}

// DeleteK3sCluster delete K3S cluster.
func (p *Amazon) DeleteK3sCluster(f bool) (err error) {
	return p.DeleteCluster(f, p.deleteInstance)
}

// SSHK3sNode ssh K3s node.
func (p *Amazon) SSHK3sNode(ip string) error {
	c := &types.Cluster{
		Metadata: p.Metadata,
		Options:  p.Options,
		Status:   p.Status,
	}
	return p.Connect(ip, &p.SSH, c, p.getInstanceNodes, p.isInstanceRunning, nil)
}

// IsClusterExist determine if the cluster exists.
func (p *Amazon) IsClusterExist() (bool, []string, error) {
	ids := make([]string, 0)

	if p.client == nil {
		p.newClient()
	}

	output, err := p.describeInstances()
	if err != nil {
		return false, ids, err
	}

	for _, instance := range output {
		if instance.State.Name != ec2types.InstanceStateNameTerminated &&
			instance.State.Name != ec2types.InstanceStateNameShuttingDown {
			ids = append(ids, *instance.InstanceId)
		}
	}

	return len(ids) > 0, ids, nil
}

// GenerateMasterExtraArgs generates K3S master extra args.
func (p *Amazon) GenerateMasterExtraArgs(cluster *types.Cluster, master types.Node) string {
	if option, ok := cluster.Options.(typesaws.Options); ok {
		if option.CloudControllerManager {
			return fmt.Sprintf(" --kubelet-arg=cloud-provider=external --kubelet-arg=provider-id=aws:///%s/%s --node-name='$(hostname -f)'", option.Zone, master.InstanceID)
		}
	}
	return ""
}

// GenerateWorkerExtraArgs generates K3S worker extra args.
func (p *Amazon) GenerateWorkerExtraArgs(cluster *types.Cluster, worker types.Node) string {
	return p.GenerateMasterExtraArgs(cluster, worker)
}

// SetOptions set options.
func (p *Amazon) SetOptions(opt []byte) error {
	sourceOption := reflect.ValueOf(&p.Options).Elem()
	option := &typesaws.Options{}
	err := json.Unmarshal(opt, option)
	if err != nil {
		return err
	}
	targetOption := reflect.ValueOf(option).Elem()
	utils.MergeConfig(sourceOption, targetOption)
	return nil
}

// GetCluster returns cluster status.
func (p *Amazon) GetCluster(kubecfg string) *types.ClusterInfo {
	c := &types.ClusterInfo{
		ID:       p.ContextName,
		Name:     p.Name,
		Provider: p.GetProviderName(),
		Region:   p.Region,
	}

	return p.GetClusterStatus(kubecfg, c, p.getInstanceNodes)
}

// DescribeCluster describe cluster info.
func (p *Amazon) DescribeCluster(kubecfg string) *types.ClusterInfo {
	c := &types.ClusterInfo{
		Name:     p.Name,
		Region:   p.Region,
		Zone:     p.Zone,
		Provider: p.GetProviderName(),
	}
	return p.Describe(kubecfg, c, p.getInstanceNodes)
}

// GetProviderOptions get provider options.
func (p *Amazon) GetProviderOptions(opt []byte) (interface{}, error) {
	options := &typesaws.Options{}
	err := json.Unmarshal(opt, options)
	return options, err
}

// SetConfig set cluster config.
func (p *Amazon) SetConfig(config []byte) error {
	c, err := p.SetClusterConfig(config)
	if err != nil {
		return err
	}
	sourceOption := reflect.ValueOf(&p.Options).Elem()
	b, err := json.Marshal(c.Options)
	if err != nil {
		return err
	}
	opt := &typesaws.Options{}
	err = json.Unmarshal(b, opt)
	if err != nil {
		return err
	}
	targetOption := reflect.ValueOf(opt).Elem()
	utils.MergeConfig(sourceOption, targetOption)

	return nil
}

func (p *Amazon) rollbackInstance(ids []string) error {
	return p.terminateInstance(ids)
}

func (p *Amazon) generateInstance(ssh *types.SSH) (*types.Cluster, error) {
	p.newClient()
	masterNum, _ := strconv.Atoi(p.Master)
	workerNum, _ := strconv.Atoi(p.Worker)

	p.Logger.Infof("[%s] %d masters and %d workers will be added in region %s", p.GetProviderName(), masterNum, workerNum, p.Region)

	if err := p.createKeyPair(ssh); err != nil {
		return nil, err
	}

	if p.SecurityGroup == "" {
		if err := p.configSecurityGroup(); err != nil {
			return nil, err
		}
	}

	if p.UserDataPath != "" {
		userDataBytes, err := os.ReadFile(p.UserDataPath)
		if err != nil {
			return nil, err
		}
		if len(userDataBytes) > 0 {
			p.UserDataContent = base64.StdEncoding.EncodeToString(userDataBytes)
		}
	}

	// run ecs master instances.
	if masterNum > 0 {
		p.Logger.Infof("[%s] prepare for %d of master instances", p.GetProviderName(), masterNum)
		if err := p.runInstances(masterNum, true, ssh); err != nil {
			return nil, err
		}
		p.Logger.Infof("[%s] %d of master instances created successfully", p.GetProviderName(), masterNum)
	}

	// run ecs worker instances.
	if workerNum > 0 {
		p.Logger.Infof("[%s] prepare for %d of worker instances", p.GetProviderName(), workerNum)
		if err := p.runInstances(workerNum, false, ssh); err != nil {
			return nil, err
		}
		p.Logger.Infof("[%s] %d of worker instances created successfully", p.GetProviderName(), workerNum)
	}

	if err := p.getInstanceStatus(string(ec2types.InstanceStateNameRunning)); err != nil {
		return nil, err
	}

	c := &types.Cluster{
		Metadata: p.Metadata,
		Options:  p.Options,
		Status:   p.Status,
	}
	c.ContextName = p.ContextName

	if p.CloudControllerManager {
		// generate tags for security group and subnet
		// https://docs.ranchermanager.rancher.io/v2.6/how-to-guides/new-user-guides/kubernetes-clusters-in-rancher-setup/set-up-cloud-providers/amazon
		err := p.addTagsForCCMResource()
		if err != nil {
			return nil, err
		}
		c.MasterExtraArgs += " --disable-cloud-controller --disable servicelb,traefik,local-storage"
	}
	c.SSH = *ssh

	return c, nil
}

func (p *Amazon) newClient() {
	// static credentials only, so the aws-sdk-go-v2/config default chain
	// (SSO/OIDC/STS/IMDS) is deliberately not linked in.
	cfg := aws.Config{
		Region: p.Region,
		Credentials: aws.CredentialsProviderFunc(
			func(context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     p.AccessKey,
					SecretAccessKey: p.SecretKey,
					SessionToken:    p.SessionToken,
				}, nil
			}),
	}
	p.client = ec2.NewFromConfig(cfg)
}

func (p *Amazon) runInstances(num int, master bool, ssh *types.SSH) error {
	rootSize, err := strconv.ParseInt(p.RootSize, 10, 64)
	if err != nil {
		return fmt.Errorf("[%s] --root-size is invalid %v, must be integer: %v", p.GetProviderName(), p.RootSize, err)
	}
	img, err := p.client.DescribeImages(context.TODO(), &ec2.DescribeImagesInput{
		ImageIds: stringSlice([]string{p.AMI}),
	})
	if err != nil {
		return fmt.Errorf("[%s] AMI %s is invalid: %v", p.GetProviderName(), p.AMI, err)
	}
	// get root device name from AMI
	deviceName := aws.String(defaultDeviceName)
	if len(img.Images) > 0 {
		deviceName = img.Images[0].RootDeviceName
	}
	bdm := ec2types.BlockDeviceMapping{
		DeviceName: deviceName,
		Ebs: &ec2types.EbsBlockDevice{
			VolumeSize:          int32Ptr(rootSize),
			VolumeType:          ec2types.VolumeType(p.VolumeType),
			DeleteOnTermination: aws.Bool(true),
		},
	}
	netSpecs := []ec2types.InstanceNetworkInterfaceSpecification{{
		DeviceIndex:              int32Ptr(0), // eth0
		Groups:                   stringSlice([]string{p.SecurityGroup}),
		SubnetId:                 &p.SubnetID,
		AssociatePublicIpAddress: aws.Bool(true),
	}}

	var iamProfile *ec2types.IamInstanceProfileSpecification
	if master {
		iamProfile = &ec2types.IamInstanceProfileSpecification{
			Name: &p.IamInstanceProfileForControl,
		}
	} else {
		iamProfile = &ec2types.IamInstanceProfileSpecification{
			Name: &p.IamInstanceProfileForWorker,
		}
	}

	input := &ec2.RunInstancesInput{
		ImageId:  &p.AMI,
		MinCount: int32Ptr(int64(num)),
		MaxCount: int32Ptr(int64(num)),
		Placement: &ec2types.Placement{
			AvailabilityZone: &p.Zone,
		},
		KeyName:             &p.KeypairName,
		InstanceType:        ec2types.InstanceType(p.InstanceType),
		NetworkInterfaces:   netSpecs,
		IamInstanceProfile:  iamProfile,
		BlockDeviceMappings: []ec2types.BlockDeviceMapping{bdm},
		UserData:            aws.String(p.UserDataContent),
	}
	if p.RequestSpotInstance {
		input.InstanceMarketOptions = &ec2types.InstanceMarketOptionsRequest{
			MarketType: ec2types.MarketTypeSpot,
		}
		if p.SpotPrice != "" {
			input.InstanceMarketOptions.SpotOptions = &ec2types.SpotMarketOptions{
				MaxPrice: aws.String(p.SpotPrice),
			}
		}
	}

	inst, err := p.client.RunInstances(context.TODO(), input)

	if err != nil || len(inst.Instances) != num {
		return fmt.Errorf("[%s] calling runInstances error. region: %s, zone: %s, msg: [%v]",
			p.GetProviderName(), p.Region, p.Zone, err)
	}
	instanceList := inst.Instances

	ids := make([]string, 0)
	for _, ins := range instanceList {
		ids = append(ids, aws.ToString(ins.InstanceId))
		p.M.Store(aws.ToString(ins.InstanceId),
			types.Node{Master: master,
				Current:        true,
				RollBack:       true,
				InstanceID:     aws.ToString(ins.InstanceId),
				InstanceStatus: string(ins.State.Name),
				SSH:            *ssh})
	}

	return p.setInstanceTags(master, ids, p.Tags)
}

func (p *Amazon) setInstanceTags(master bool, instanceIDs []string, additionalTags []string) error {
	tags := []ec2types.Tag{
		{
			Key:   aws.String("autok3s"),
			Value: aws.String("true"),
		},
		{
			Key:   aws.String("cluster"),
			Value: aws.String(common.TagClusterPrefix + p.ContextName),
		},
		{
			Key:   aws.String("master"),
			Value: aws.String(strconv.FormatBool(master)),
		},
	}

	for _, v := range additionalTags {
		ss := strings.Split(v, "=")
		if len(ss) != 2 {
			return fmt.Errorf("tags %s invalid", v)
		}
		tags = append(tags, ec2types.Tag{
			Key:   aws.String(ss[0]),
			Value: aws.String(ss[1]),
		})
	}

	if master {
		tags = append(tags, ec2types.Tag{
			Key:   aws.String("Name"),
			Value: aws.String(fmt.Sprintf(common.MasterInstanceName, p.ContextName)),
		})
	} else {
		tags = append(tags, ec2types.Tag{
			Key:   aws.String("Name"),
			Value: aws.String(fmt.Sprintf(common.WorkerInstanceName, p.ContextName)),
		})
	}

	if p.CloudControllerManager {
		tags = append(tags, ec2types.Tag{
			Key:   aws.String(fmt.Sprintf("kubernetes.io/cluster/%s", p.ContextName)),
			Value: aws.String("owned"),
		})
	}

	_, err := p.client.CreateTags(context.TODO(), &ec2.CreateTagsInput{
		Resources: instanceIDs,
		Tags:      tags,
	})
	return err
}

func (p *Amazon) getInstanceStatus(aimStatus string) error {
	ids := make([]string, 0)
	p.M.Range(func(key, value interface{}) bool {
		ids = append(ids, key.(string))
		return true
	})

	if len(ids) > 0 {
		p.Logger.Infof("[%s] waiting for the instances %s to be in `%s` status...", p.GetProviderName(), ids, aimStatus)
		waiter := ec2.NewInstanceRunningWaiter(p.client)
		err := waiter.Wait(context.TODO(), &ec2.DescribeInstancesInput{
			InstanceIds: stringSlice(ids),
		}, 15*time.Minute)
		if err != nil {
			return err
		}

		instances, err := p.client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
			InstanceIds: stringSlice(ids),
		})
		if err != nil {
			return err
		}
		for _, reservation := range instances.Reservations {
			for _, status := range reservation.Instances {
				if string(status.State.Name) == aimStatus {
					if value, ok := p.M.Load(aws.ToString(status.InstanceId)); ok {
						v := value.(types.Node)
						v.InstanceStatus = aimStatus
						v.InternalIPAddress = []string{aws.ToString(status.PrivateIpAddress)}
						v.PublicIPAddress = []string{aws.ToString(status.PublicIpAddress)}
						v.LocalHostname = aws.ToString(status.PrivateDnsName)
						p.M.Store(aws.ToString(status.InstanceId), v)
					}
					continue
				}
			}
		}
	}

	p.Logger.Infof("[%s] instances %s are in `%s` status", p.GetProviderName(), ids, aimStatus)

	return nil
}

func (p *Amazon) getInstanceNodes() ([]types.Node, error) {
	if p.client == nil {
		p.newClient()
	}
	output, err := p.describeInstances()
	if err != nil || len(output) == 0 {
		return nil, fmt.Errorf("[%s] there's no instance for cluster %s: %v", p.GetProviderName(), p.ContextName, err)
	}
	nodes := make([]types.Node, 0)
	for _, instance := range output {
		if instance.State.Name == ec2types.InstanceStateNameTerminated {
			continue
		}
		master := false
		for _, tag := range instance.Tags {
			if strings.EqualFold(*tag.Key, "master") && strings.EqualFold(*tag.Value, "true") {
				master = true
				break
			}
		}
		nodes = append(nodes, types.Node{
			Master:            master,
			RollBack:          false,
			InstanceID:        aws.ToString(instance.InstanceId),
			InstanceStatus:    string(instance.State.Name),
			InternalIPAddress: []string{aws.ToString(instance.PrivateIpAddress)},
			PublicIPAddress:   []string{aws.ToString(instance.PublicIpAddress)}})
	}
	return nodes, nil
}

func (p *Amazon) syncInstances() error {
	output, err := p.describeInstances()
	if err != nil || len(output) == 0 {
		return fmt.Errorf("[%s] there's no instance for cluster %s: %v", p.GetProviderName(), p.Name, err)
	}
	for _, instance := range output {
		if instance.State.Name == ec2types.InstanceStateNameTerminated {
			continue
		}
		if value, ok := p.M.Load(aws.ToString(instance.InstanceId)); ok {
			v := value.(types.Node)
			v.InternalIPAddress = []string{aws.ToString(instance.PrivateIpAddress)}
			v.PublicIPAddress = []string{aws.ToString(instance.PublicIpAddress)}
			p.M.Store(aws.ToString(instance.InstanceId), v)
			continue
		}
		master := false
		for _, tag := range instance.Tags {
			if strings.EqualFold(*tag.Key, "master") && strings.EqualFold(*tag.Value, "true") {
				master = true
				break
			}
		}
		p.M.Store(aws.ToString(instance.InstanceId), types.Node{
			Master:            master,
			RollBack:          false,
			Current:           false,
			InstanceID:        aws.ToString(instance.InstanceId),
			InstanceStatus:    string(instance.State.Name),
			InternalIPAddress: []string{aws.ToString(instance.PrivateIpAddress)},
			PublicIPAddress:   []string{aws.ToString(instance.PublicIpAddress)}})
	}
	return nil
}

func (p *Amazon) describeInstances() ([]ec2types.Instance, error) {
	describeInput := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag:autok3s"),
				Values: stringSlice([]string{"true"}),
			},
			{
				Name:   aws.String("tag:cluster"),
				Values: stringSlice([]string{common.TagClusterPrefix + p.ContextName}),
			},
		},
		MaxResults: int32Ptr(int64(50)),
	}

	instanceList := make([]ec2types.Instance, 0)
	for {
		output, err := p.client.DescribeInstances(context.TODO(), describeInput)
		if output == nil || err != nil {
			if ae, ok := asAPIError(err); ok {
				if ae.ErrorCode() == "AuthFailure" {
					return nil, fmt.Errorf("[%s] invalid credential: %s", p.GetProviderName(), ae.ErrorMessage())
				}
			}
			return nil, fmt.Errorf("[%s] failed to get instance for cluster %s: %v", p.GetProviderName(), p.ContextName, err)
		}

		if len(output.Reservations) > 0 {
			for _, reservation := range output.Reservations {
				instanceList = append(instanceList, reservation.Instances...)
			}
		}
		if aws.ToString(output.NextToken) == "" {
			break
		}
		describeInput.NextToken = output.NextToken
	}
	return instanceList, nil
}

// CreateCheck check create command and flags.
func (p *Amazon) CreateCheck() error {
	if err := p.CheckCreateArgs(p.IsClusterExist); err != nil {
		return err
	}

	if err := p.ValidateRequireSSHPrivateKey(); p.KeypairName != "" && err != nil {
		return fmt.Errorf("[%s] calling preflight error: %s with --keypair-name %s", p.GetProviderName(), err.Error(), p.KeypairName)
	}

	masterNum, err := strconv.Atoi(p.Master)
	if masterNum < 1 || err != nil {
		return fmt.Errorf("[%s] calling preflight error: `--master` number must >= 1",
			p.GetProviderName())
	}

	if masterNum > 0 && p.CloudControllerManager && p.IamInstanceProfileForControl == "" {
		return fmt.Errorf("[%s] calling preflight error: need to set `--iam-instance-profile-control` if enabled Amazon Cloud Controller Manager", p.GetProviderName())
	}

	workerNum, err := strconv.Atoi(p.Worker)
	if err != nil {
		return fmt.Errorf("[%s] calling preflight error: `--worker` must be number",
			p.GetProviderName())
	}
	if workerNum > 0 && p.CloudControllerManager && p.IamInstanceProfileForWorker == "" {
		return fmt.Errorf("[%s] calling preflight error: need to set `--iam-instance-profile-worker` if enabled Amazon Cloud Controller Manager", p.GetProviderName())
	}

	// check key pair.
	if p.KeypairName != "" {
		_, err = p.client.DescribeKeyPairs(context.TODO(), &ec2.DescribeKeyPairsInput{
			KeyNames: []string{p.KeypairName},
		})
		if err != nil {
			if ae, ok := asAPIError(err); ok {
				if ae.ErrorCode() == "AuthFailure" {
					return fmt.Errorf("[%s] invalid credential: %s", p.GetProviderName(), ae.ErrorMessage())
				}
			}
			return fmt.Errorf("[%s] failed to get keypair by name %s, got error: %v", p.GetProviderName(), p.KeypairName, err)
		}
	}

	// check vpc and subnet.
	if p.VpcID == "" {
		p.VpcID, err = p.getDefaultVPCId()
		if err != nil {
			return err
		}
	}

	if p.SubnetID == "" && p.VpcID == "" {
		return fmt.Errorf("[%s] calling preflight error: can't generate instance without vpc and subnet", p.GetProviderName())
	}

	if p.SubnetID != "" && p.VpcID != "" {
		// check subnet is belongs to vpc.
		subnetFilter := []ec2types.Filter{
			{
				Name:   aws.String("subnet-id"),
				Values: []string{p.SubnetID},
			},
		}

		subnets, err := p.client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{
			Filters: subnetFilter,
		})
		if err != nil {
			return err
		}

		if subnets == nil || len(subnets.Subnets) == 0 {
			return fmt.Errorf("[%s] there's not subnet found by id %s", p.GetProviderName(), p.SubnetID)
		}

		if aws.ToString(subnets.Subnets[0].VpcId) != p.VpcID {
			return fmt.Errorf("[%s] subnetId %s does not belong to VpcId: %s", p.GetProviderName(), p.SubnetID, p.VpcID)
		}
	}

	if p.SubnetID == "" {
		filters := []ec2types.Filter{
			{
				Name:   aws.String("availability-zone"),
				Values: []string{p.Zone},
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []string{p.VpcID},
			},
		}

		subnets, err := p.client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{
			Filters: filters,
		})
		if err != nil {
			return err
		}

		if len(subnets.Subnets) == 0 {
			return fmt.Errorf("can't get subnets for vpc %s at zone %s", p.VpcID, p.Zone)
		}

		// find default subnet.
		if len(subnets.Subnets) > 1 {
			for _, subnet := range subnets.Subnets {
				if subnet.DefaultForAz != nil && *subnet.DefaultForAz {
					p.SubnetID = *subnet.SubnetId
					break
				}
			}
		}

		if p.SubnetID == "" {
			p.SubnetID = *subnets.Subnets[0].SubnetId
		}
	}
	// check user-data
	if p.UserDataPath != "" {
		if _, err := os.Stat(p.UserDataPath); err != nil {
			return err
		}
	}

	return nil
}

// JoinCheck check join command and flags.
func (p *Amazon) JoinCheck() error {
	if err := p.CheckJoinArgs(p.IsClusterExist); err != nil {
		return err
	}
	masterNum, err := strconv.Atoi(p.Master)
	if err != nil {
		return fmt.Errorf("[%s] calling preflight error: `--master` must be number",
			p.GetProviderName())
	}
	workerNum, err := strconv.Atoi(p.Worker)
	if err != nil {
		return fmt.Errorf("[%s] calling preflight error: `--worker` must be number",
			p.GetProviderName())
	}

	if masterNum > 0 && p.CloudControllerManager && p.IamInstanceProfileForControl == "" {
		return fmt.Errorf("[%s] calling preflight error: need to set `--iam-instance-profile-control` if enabled Amazon Cloud Controller Manager", p.GetProviderName())
	}

	if workerNum > 0 && p.CloudControllerManager && p.IamInstanceProfileForWorker == "" {
		return fmt.Errorf("[%s] calling preflight error: need to set `--iam-instance-profile-worker` if enabled Amazon Cloud Controller Manager", p.GetProviderName())
	}
	return nil
}

func (p *Amazon) createKeyPair(ssh *types.SSH) error {
	if p.KeypairName != "" && ssh.SSHKeyPath == "" && p.KeypairName != p.ContextName {
		return fmt.Errorf("[%s] calling preflight error: --ssh-key-path must set with --key-pair %s", p.GetProviderName(), p.KeypairName)
	}

	// check create & upload keypair.
	if ssh.SSHKeyPath == "" {
		if _, err := os.Stat(common.GetDefaultSSHKeyPath(p.ContextName, p.GetProviderName())); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			pk, err := putil.CreateKeyPair(ssh, p.GetProviderName(), p.ContextName, p.KeypairName)
			if err != nil {
				return err
			}

			if pk != nil {
				keyName := p.ContextName
				p.Logger.Infof("[%s] creating key pair %s...", p.GetProviderName(), keyName)
				_, err = p.client.ImportKeyPair(context.TODO(), &ec2.ImportKeyPairInput{
					KeyName:           &keyName,
					PublicKeyMaterial: pk,
				})
				if err != nil {
					if ae, ok := asAPIError(err); ok && strings.EqualFold(ae.ErrorCode(), "InvalidKeyPair.Duplicate") {
						// atomicity preservation:
						//   1. delete id_rsa & id_rsa.pub.
						//   2. reset cluster's ssh-key-path'.
						p.SSHKeyPath = ""
						_ = utils.RemoveSSHKey(common.GetDefaultSSHKeyPath(p.ContextName, p.GetProviderName()))
						_ = utils.RemoveSSHKey(common.GetDefaultSSHPublicKeyPath(p.ContextName, p.GetProviderName()))
						return fmt.Errorf("[%s] key pair %s is duplicate on aws, please create new one or use the one which already exist on %s", p.GetProviderName(), keyName, p.GetProviderName())
					}
					return fmt.Errorf("[%s] created key pair %s error: %w", p.GetProviderName(), keyName, err)
				}
				p.KeypairName = keyName
				p.Logger.Infof("[%s] successfully created key pair %s", p.GetProviderName(), keyName)
			}
		}
		p.KeypairName = p.ContextName
		ssh.SSHKeyPath = common.GetDefaultSSHKeyPath(p.ContextName, p.GetProviderName())
	}

	return nil
}

func (p *Amazon) getDefaultVPCId() (string, error) {
	output, err := p.client.DescribeAccountAttributes(context.TODO(), &ec2.DescribeAccountAttributesInput{})
	if err != nil {
		if ae, ok := asAPIError(err); ok {
			if ae.ErrorCode() == "AuthFailure" {
				return "", fmt.Errorf("[%s] invalid credential: %s", p.GetProviderName(), ae.ErrorMessage())
			}
		}
		return "", err
	}

	for _, attribute := range output.AccountAttributes {
		if aws.ToString(attribute.AttributeName) == "default-vpc" {
			value := aws.ToString(attribute.AttributeValues[0].AttributeValue)
			if value == "none" {
				return "", errors.New("there's 'none' for default vpc")
			}
			return value, nil
		}
	}

	return "", errors.New("couldn't get default vpc")
}

func (p *Amazon) configSecurityGroup() error {
	p.Logger.Infof("[%s] config default security group for %s in region %s", p.GetProviderName(), p.VpcID, p.Region)

	filters := []ec2types.Filter{
		{
			Name:   aws.String("group-name"),
			Values: stringSlice([]string{defaultSecurityGroupName}),
		},
		{
			Name:   aws.String("vpc-id"),
			Values: stringSlice([]string{p.VpcID}),
		},
	}
	groups, err := p.client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
		Filters: filters,
	})
	if err != nil {
		if ae, ok := asAPIError(err); ok {
			if ae.ErrorCode() == "AuthFailure" {
				return fmt.Errorf("[%s] invalid credential: %s", p.GetProviderName(), ae.ErrorMessage())
			}
		}
		return err
	}

	var securityGroup ec2types.SecurityGroup
	securityGroupFound := false
	if len(groups.SecurityGroups) > 0 {
		// get default security group.
		securityGroup = groups.SecurityGroups[0]
		securityGroupFound = true
	}

	if !securityGroupFound {
		p.Logger.Infof("creating security group (%s) in %s", defaultSecurityGroupName, p.VpcID)
		groupResp, err := p.client.CreateSecurityGroup(context.TODO(), &ec2.CreateSecurityGroupInput{
			GroupName:   aws.String(defaultSecurityGroupName),
			Description: aws.String("default security group generated by autok3s"),
			VpcId:       aws.String(p.VpcID),
		})
		if err != nil {
			return err
		}
		// Manually translate into the security group construct.
		securityGroup = ec2types.SecurityGroup{
			GroupId:   groupResp.GroupId,
			VpcId:     aws.String(p.VpcID),
			GroupName: aws.String(defaultSecurityGroupName),
		}
		// wait until created (dat eventual consistency).
		p.Logger.Infof("waiting for group (%s) to become available", *securityGroup.GroupId)
		err = utils.WaitFor(func() (bool, error) {
			s, err := p.getSecurityGroup(groupResp.GroupId)
			if s != nil && err == nil {
				return true, nil
			}
			return false, err
		})
		if err != nil {
			return err
		}
	}
	p.SecurityGroup = aws.ToString(securityGroup.GroupId)
	permissionList := p.configPermission(securityGroup)
	if len(permissionList) != 0 {
		p.Logger.Infof("authorizing group %s with permissions: %v", defaultSecurityGroupName, permissionList)
		_, err := p.client.AuthorizeSecurityGroupIngress(context.TODO(), &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       securityGroup.GroupId,
			IpPermissions: permissionList,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Amazon) getSecurityGroup(id *string) (*ec2types.SecurityGroup, error) {
	securityGroups, err := p.client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{aws.ToString(id)},
	})
	if err != nil {
		return nil, err
	}
	if len(securityGroups.SecurityGroups) == 0 {
		return nil, fmt.Errorf("[%s] security group %s not found", p.GetProviderName(), aws.ToString(id))
	}
	return &securityGroups.SecurityGroups[0], nil
}

func (p *Amazon) configPermission(group ec2types.SecurityGroup) []ec2types.IpPermission {
	perms := make([]ec2types.IpPermission, 0)
	hasPorts := make(map[string]bool)
	for _, p := range group.IpPermissions {
		if p.FromPort != nil {
			hasPorts[fmt.Sprintf("%d/%s", *p.FromPort, *p.IpProtocol)] = true
		}
	}

	if !hasPorts["22/tcp"] {
		perms = append(perms, ec2types.IpPermission{
			IpProtocol: aws.String("tcp"),
			FromPort:   int32Ptr(int64(22)),
			ToPort:     int32Ptr(int64(22)),
			IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(ipRange)}},
		})
	}

	if !hasPorts["6443/tcp"] {
		perms = append(perms, ec2types.IpPermission{
			IpProtocol: aws.String("tcp"),
			FromPort:   int32Ptr(int64(6443)),
			ToPort:     int32Ptr(int64(6443)),
			IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(ipRange)}},
		})
	}

	if !hasPorts["10250/tcp"] {
		perms = append(perms, ec2types.IpPermission{
			IpProtocol: aws.String("tcp"),
			FromPort:   int32Ptr(int64(10250)),
			ToPort:     int32Ptr(int64(10250)),
			IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(ipRange)}},
		})
	}

	if (p.Network == "" || p.Network == "vxlan") && !hasPorts["8472/udp"] {
		if !hasPorts["8472/udp"] {
			// udp 8472 for flannel vxlan.
			perms = append(perms, ec2types.IpPermission{
				IpProtocol: aws.String("udp"),
				FromPort:   int32Ptr(int64(8472)),
				ToPort:     int32Ptr(int64(8472)),
				IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(ipRange)}},
			})
		}
	}

	if p.Cluster && (!hasPorts["2379/tcp"] || !hasPorts["2380/tcp"]) {
		cidr, err := p.getSubnetCIDR()
		if err != nil || cidr == "" {
			p.Logger.Errorf("[%s] failed to get subnet cidr with id %s, error: %v", p.GetProviderName(), p.SubnetID, err)
			cidr = ipRange
		}
		perms = append(perms, ec2types.IpPermission{
			IpProtocol: aws.String("tcp"),
			FromPort:   int32Ptr(int64(2379)),
			ToPort:     int32Ptr(int64(2379)),
			IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(cidr)}},
		})
		perms = append(perms, ec2types.IpPermission{
			IpProtocol: aws.String("tcp"),
			FromPort:   int32Ptr(int64(2380)),
			ToPort:     int32Ptr(int64(2380)),
			IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(cidr)}},
		})
	}

	return perms
}

func (p *Amazon) deleteInstance(f bool) (string, error) {
	p.newClient()
	p.GenerateClusterName()
	exist, ids, err := p.IsClusterExist()
	if err != nil {
		return "", fmt.Errorf("[%s] calling describe instance error, msg: %v", p.GetProviderName(), err)
	}
	if !exist {
		p.Logger.Errorf("[%s] cluster %s is not exist", p.GetProviderName(), p.Name)
		if !f {
			return "", fmt.Errorf("[%s] calling preflight error: cluster name `%s` do not exist", p.GetProviderName(), p.Name)
		}
		return p.ContextName, nil
	}

	// CCM will create elb for kubernetes service
	if p.CloudControllerManager {
		p.Logger.Warnf("[%s] Please ensure all services has released before remove the cluster, if not, please check the AWS ELB to ensure all ELB has been released.", p.GetProviderName())
	}
	if err = p.terminateInstance(ids); err != nil {
		return "", err
	}
	p.Logger.Infof("[%s] successfully terminate instances for cluster %s", p.GetProviderName(), p.Name)
	return p.ContextName, nil
}

func (p *Amazon) getSubnetCIDR() (string, error) {
	subnetFilter := []ec2types.Filter{
		{
			Name:   aws.String("subnet-id"),
			Values: []string{p.SubnetID},
		},
	}

	subnets, err := p.client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{
		Filters: subnetFilter,
	})
	if err != nil {
		return "", err
	}

	if subnets == nil || len(subnets.Subnets) == 0 {
		return "", fmt.Errorf("[%s] there's not subnet found by id %s", p.GetProviderName(), p.SubnetID)
	}

	return aws.ToString(subnets.Subnets[0].CidrBlock), nil
}

func (p *Amazon) addTagsForCCMResource() error {
	// get security group
	result, err := p.client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
		GroupIds: stringSlice([]string{p.SecurityGroup}),
	})
	if err != nil || result == nil || len(result.SecurityGroups) == 0 {
		return fmt.Errorf("[%s] failed to get security group %s with error: %v", p.GetProviderName(), p.SecurityGroup, err)
	}
	tags := result.SecurityGroups[0].Tags
	tags = append(tags, ec2types.Tag{
		Key:   aws.String(fmt.Sprintf("kubernetes.io/cluster/%s", p.ContextName)),
		Value: aws.String("shared"),
	})
	_, err = p.client.CreateTags(context.TODO(), &ec2.CreateTagsInput{
		Resources: stringSlice([]string{p.SecurityGroup}),
		Tags:      tags,
	})
	if err != nil {
		return err
	}

	// get subnet.
	subnetFilter := []ec2types.Filter{
		{
			Name:   aws.String("subnet-id"),
			Values: []string{p.SubnetID},
		},
	}
	subnets, err := p.client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{
		Filters: subnetFilter,
	})
	if err != nil || subnets == nil || len(subnets.Subnets) == 0 {
		return fmt.Errorf("[%s] failed to get subnets %s, error: %v", p.GetProviderName(), p.SubnetID, err)
	}
	subnetTags := subnets.Subnets[0].Tags
	subnetTags = append(subnetTags, ec2types.Tag{
		Key:   aws.String(fmt.Sprintf("kubernetes.io/cluster/%s", p.ContextName)),
		Value: aws.String("shared"),
	})
	_, err = p.client.CreateTags(context.TODO(), &ec2.CreateTagsInput{
		Resources: stringSlice([]string{p.SubnetID}),
		Tags:      subnetTags,
	})

	return err
}

func (p *Amazon) removeTagsForCCMResource() error {
	deletedTags := []ec2types.Tag{
		{
			Key:   aws.String(fmt.Sprintf("kubernetes.io/cluster/%s", p.ContextName)),
			Value: aws.String("shared"),
		},
	}
	_, err := p.client.DeleteTags(context.TODO(), &ec2.DeleteTagsInput{
		Resources: stringSlice([]string{p.SecurityGroup}),
		Tags:      deletedTags,
	})
	if err != nil {
		return err
	}
	_, err = p.client.DeleteTags(context.TODO(), &ec2.DeleteTagsInput{
		Resources: stringSlice([]string{p.SubnetID}),
		Tags:      deletedTags,
	})
	return err
}

func (p *Amazon) terminateInstance(ids []string) error {
	if len(ids) > 0 {
		tags := []ec2types.Tag{
			{
				Key:   aws.String("autok3s"),
				Value: aws.String("true"),
			},
			{
				Key:   aws.String("cluster"),
				Value: aws.String(common.TagClusterPrefix + p.ContextName),
			},
		}
		tagInput := &ec2.DeleteTagsInput{
			Tags:      tags,
			Resources: stringSlice(ids),
		}
		_, err := p.client.DeleteTags(context.TODO(), tagInput)
		if err != nil {
			return err
		}
		p.Logger.Infof("[%s] terminate instance %v", p.GetProviderName(), ids)
		input := &ec2.TerminateInstancesInput{
			InstanceIds: stringSlice(ids),
		}
		_, err = p.client.TerminateInstances(context.TODO(), input)
		if err != nil {
			return err
		}
		if p.CloudControllerManager {
			err = p.removeTagsForCCMResource()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Amazon) isInstanceRunning(state string) bool {
	return state == string(ec2types.InstanceStateNameRunning)
}

func getAWSCCMManifest(k3sversion, cidr string) string {
	ccmVersion, err := getCCMVersion(k3sversion)
	if err != nil {
		logrus.Warnf("failed to get CCM version for k3s version %v, skip generating CCM manifest, %v", k3sversion, err)
		return ""
	}
	extraArgs, err := getCCMExtraArgs(k3sversion)
	if err != nil {
		logrus.Warnf("failed to get CCM extra args for k3s version %v, assuming no extra args, %v", k3sversion, err)
		extraArgs = []string{}
	}

	rtn := bytes.NewBuffer([]byte{})
	if err := ccmTemplate.Execute(rtn, map[string]interface{}{
		"Version":     ccmVersion,
		"ExtraArgs":   extraArgs,
		"ClusterCIDR": cidr,
	}); err != nil {
		logrus.Warnf("failed to execute AWS CCM template, assuming no manifest, %v", err)
	}

	return rtn.String()
}

func getCCMVersion(k3sversion string) (string, error) {
	if k3sversion != "" {
		versionSemver, err := semver.NewVersion(k3sversion)
		if err != nil {
			return "", err
		}
		for version, targetVersion := range ccmVersionMap {
			// assume the constrant is validated
			if cst, _ := semver.NewConstraint(version); cst.Check(versionSemver) {
				return targetVersion, nil
			}
		}
	}

	// should never meet this condition because the version map covers all k8s version.
	return ccmVersionMap[">= 1.27"], nil
}

func getCCMExtraArgs(k3sversion string) ([]string, error) {
	if k3sversion == "" {
		return ccmExtraArgs[">= 1.21"], nil
	}
	versionSemver, err := semver.NewVersion(k3sversion)
	if err != nil {
		return nil, err
	}
	for version := range ccmExtraArgs {
		// assume the constrant is validated
		if cst, _ := semver.NewConstraint(version); cst.Check(versionSemver) {
			return ccmExtraArgs[version], nil
		}
	}
	return []string{}, nil
}
