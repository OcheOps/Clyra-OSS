package props

import (
	"encoding/json"
	"fmt"

	"github.com/oche/diagram/internal/iac/graph"
)

func init() {
	graph.ValidateProps = func(kind graph.ResourceKind, props map[string]any) error {
		fn, ok := validators[kind]
		if !ok {
			return fmt.Errorf("no validator registered for kind %s", kind)
		}
		return fn(props)
	}
}

var validators = map[graph.ResourceKind]func(map[string]any) error{
	graph.KindVPC:                     validateVPC,
	graph.KindSubnet:                  validateSubnet,
	graph.KindInternetGateway:         func(map[string]any) error { return nil },
	graph.KindRouteTable:              validateRouteTable,
	graph.KindSecurityGroup:           validateSecurityGroup,
	graph.KindALB:                     validateALB,
	graph.KindTargetGroup:             validateTargetGroup,
	graph.KindListener:                validateListener,
	graph.KindECSCluster:              func(map[string]any) error { return nil },
	graph.KindECSService:              validateECSService,
	graph.KindECSTaskDefinition:       validateECSTaskDefinition,
	graph.KindRDSInstance:             validateRDSInstance,
	graph.KindRDSSubnetGroup:          validateRDSSubnetGroup,
	graph.KindElastiCacheCluster:      validateElastiCacheCluster,
	graph.KindElastiCacheSubnetGroup:  validateElastiCacheSubnetGroup,
	graph.KindLambdaFunction:          validateLambdaFunction,
	graph.KindLambdaPermission:        validateLambdaPermission,
	graph.KindAPIGatewayV2API:         validateAPIGatewayV2API,
	graph.KindAPIGatewayV2Route:       validateAPIGatewayV2Route,
	graph.KindAPIGatewayV2Integration: validateAPIGatewayV2Integration,
	graph.KindDynamoDBTable:           validateDynamoDBTable,
	graph.KindS3Bucket:                validateS3Bucket,
	graph.KindOpenSearchDomain:        validateOpenSearchDomain,
	graph.KindEC2Instance:             validateEC2Instance,
	graph.KindCloudFrontDist:          validateCloudFrontDist,
	graph.KindIAMRole:                 validateIAMRole,
	graph.KindCloudWatchLogGroup:      validateCloudWatchLogGroup,
	graph.KindNATGateway:              validateNATGateway,
}

type VPC struct {
	CIDRBlock          string `json:"cidr_block"`
	EnableDNSHostnames bool   `json:"enable_dns_hostnames"`
}

func AsVPC(m map[string]any) (*VPC, error) { return decode[VPC](m) }
func validateVPC(m map[string]any) error {
	v, err := AsVPC(m)
	if err != nil {
		return err
	}
	if v.CIDRBlock == "" {
		return fmt.Errorf("cidr_block required")
	}
	return nil
}

type Subnet struct {
	VPC                 string `json:"vpc"`
	CIDRBlock           string `json:"cidr_block"`
	AvailabilityZone    string `json:"availability_zone"`
	MapPublicIPOnLaunch bool   `json:"map_public_ip_on_launch"`
}

func AsSubnet(m map[string]any) (*Subnet, error) { return decode[Subnet](m) }
func validateSubnet(m map[string]any) error {
	s, err := AsSubnet(m)
	if err != nil {
		return err
	}
	if s.VPC == "" || s.CIDRBlock == "" || s.AvailabilityZone == "" {
		return fmt.Errorf("vpc, cidr_block, and availability_zone are required")
	}
	return nil
}

type NATGateway struct {
	Subnet string `json:"subnet"`
}

func AsNATGateway(m map[string]any) (*NATGateway, error) { return decode[NATGateway](m) }
func validateNATGateway(m map[string]any) error {
	n, err := AsNATGateway(m)
	if err != nil {
		return err
	}
	if n.Subnet == "" {
		return fmt.Errorf("subnet required")
	}
	return nil
}

type RouteTable struct {
	VPC string `json:"vpc"`
}

func AsRouteTable(m map[string]any) (*RouteTable, error) { return decode[RouteTable](m) }
func validateRouteTable(m map[string]any) error {
	r, err := AsRouteTable(m)
	if err != nil {
		return err
	}
	if r.VPC == "" {
		return fmt.Errorf("vpc required")
	}
	return nil
}

type SecurityGroupRule struct {
	FromPort int    `json:"from_port"`
	ToPort   int    `json:"to_port"`
	Protocol string `json:"protocol"`
	CIDR     string `json:"cidr,omitempty"`
	SG       string `json:"sg,omitempty"`
}

type SecurityGroup struct {
	VPC           string              `json:"vpc"`
	Ingress       []SecurityGroupRule `json:"ingress,omitempty"`
	IngressFromSG []SecurityGroupRule `json:"ingress_from_sg,omitempty"`
}

func AsSecurityGroup(m map[string]any) (*SecurityGroup, error) { return decode[SecurityGroup](m) }
func validateSecurityGroup(m map[string]any) error {
	s, err := AsSecurityGroup(m)
	if err != nil {
		return err
	}
	if s.VPC == "" {
		return fmt.Errorf("vpc required")
	}
	if len(s.Ingress) == 0 && len(s.IngressFromSG) == 0 {
		return fmt.Errorf("at least one ingress or ingress_from_sg rule is required")
	}
	return nil
}

type ALB struct {
	Subnets        []string `json:"subnets"`
	SecurityGroups []string `json:"security_groups"`
}

func AsALB(m map[string]any) (*ALB, error) { return decode[ALB](m) }
func validateALB(m map[string]any) error {
	a, err := AsALB(m)
	if err != nil {
		return err
	}
	if len(a.Subnets) < 2 || len(a.SecurityGroups) == 0 {
		return fmt.Errorf("subnets (>=2) and security_groups are required")
	}
	return nil
}

type TargetGroup struct {
	VPC        string `json:"vpc"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	TargetType string `json:"target_type"`
}

func AsTargetGroup(m map[string]any) (*TargetGroup, error) { return decode[TargetGroup](m) }
func validateTargetGroup(m map[string]any) error {
	t, err := AsTargetGroup(m)
	if err != nil {
		return err
	}
	if t.VPC == "" || t.Port == 0 || t.Protocol == "" || t.TargetType == "" {
		return fmt.Errorf("vpc, port, protocol, and target_type are required")
	}
	return nil
}

type Listener struct {
	LoadBalancer string `json:"load_balancer"`
	Port         int    `json:"port"`
	Protocol     string `json:"protocol"`
	TargetGroup  string `json:"target_group"`
}

func AsListener(m map[string]any) (*Listener, error) { return decode[Listener](m) }
func validateListener(m map[string]any) error {
	l, err := AsListener(m)
	if err != nil {
		return err
	}
	if l.LoadBalancer == "" || l.Port == 0 || l.Protocol == "" || l.TargetGroup == "" {
		return fmt.Errorf("load_balancer, port, protocol, and target_group are required")
	}
	return nil
}

type ECSTaskDefinition struct {
	Family        string `json:"family"`
	CPU           string `json:"cpu"`
	Memory        string `json:"memory"`
	Image         string `json:"image"`
	ContainerName string `json:"container_name,omitempty"`
	ContainerPort int    `json:"container_port"`
}

func AsECSTaskDefinition(m map[string]any) (*ECSTaskDefinition, error) {
	return decode[ECSTaskDefinition](m)
}
func validateECSTaskDefinition(m map[string]any) error {
	t, err := AsECSTaskDefinition(m)
	if err != nil {
		return err
	}
	if t.Family == "" || t.CPU == "" || t.Memory == "" || t.Image == "" || t.ContainerPort == 0 {
		return fmt.Errorf("family, cpu, memory, image, and container_port are required")
	}
	return nil
}

type ECSService struct {
	Cluster        string   `json:"cluster"`
	TaskDefinition string   `json:"task_definition"`
	DesiredCount   int      `json:"desired_count"`
	LaunchType     string   `json:"launch_type"`
	Subnets        []string `json:"subnets"`
	SecurityGroups []string `json:"security_groups"`
	TargetGroup    string   `json:"target_group,omitempty"`
	ContainerName  string   `json:"container_name,omitempty"`
	ContainerPort  int      `json:"container_port,omitempty"`
	AssignPublicIP bool     `json:"assign_public_ip"`
}

func AsECSService(m map[string]any) (*ECSService, error) { return decode[ECSService](m) }
func validateECSService(m map[string]any) error {
	s, err := AsECSService(m)
	if err != nil {
		return err
	}
	if s.Cluster == "" || s.TaskDefinition == "" {
		return fmt.Errorf("cluster and task_definition are required")
	}
	if s.LaunchType != "" && s.LaunchType != "FARGATE" && s.LaunchType != "EC2" {
		return fmt.Errorf("launch_type must be FARGATE or EC2")
	}
	if len(s.Subnets) == 0 || len(s.SecurityGroups) == 0 {
		return fmt.Errorf("subnets and security_groups are required")
	}
	return nil
}

type RDSInstance struct {
	Engine            string   `json:"engine"`
	EngineVersion     string   `json:"engine_version"`
	InstanceClass     string   `json:"instance_class"`
	AllocatedStorage  int      `json:"allocated_storage"`
	DBName            string   `json:"db_name"`
	DBSubnetGroup     string   `json:"db_subnet_group"`
	VPCSecurityGroups []string `json:"vpc_security_groups"`
	Port              int      `json:"port,omitempty"`
	SkipFinalSnapshot bool     `json:"skip_final_snapshot"`
	StorageEncrypted  bool     `json:"storage_encrypted"`
}

func AsRDSInstance(m map[string]any) (*RDSInstance, error) { return decode[RDSInstance](m) }
func validateRDSInstance(m map[string]any) error {
	r, err := AsRDSInstance(m)
	if err != nil {
		return err
	}
	validEngines := map[string]bool{"postgres": true, "mysql": true, "mariadb": true}
	if !validEngines[r.Engine] {
		return fmt.Errorf("engine %q not in {postgres, mysql, mariadb}", r.Engine)
	}
	if r.InstanceClass == "" {
		return fmt.Errorf("instance_class required")
	}
	if r.AllocatedStorage < 20 {
		return fmt.Errorf("allocated_storage must be >= 20 GiB")
	}
	if r.DBSubnetGroup == "" {
		return fmt.Errorf("db_subnet_group required")
	}
	return nil
}

type RDSSubnetGroup struct {
	Subnets []string `json:"subnets"`
}

func AsRDSSubnetGroup(m map[string]any) (*RDSSubnetGroup, error) { return decode[RDSSubnetGroup](m) }
func validateRDSSubnetGroup(m map[string]any) error {
	g, err := AsRDSSubnetGroup(m)
	if err != nil {
		return err
	}
	if len(g.Subnets) < 2 {
		return fmt.Errorf("subnets must contain at least two subnet IDs")
	}
	return nil
}

type ElastiCacheSubnetGroup struct {
	Subnets []string `json:"subnets"`
}

func AsElastiCacheSubnetGroup(m map[string]any) (*ElastiCacheSubnetGroup, error) {
	return decode[ElastiCacheSubnetGroup](m)
}
func validateElastiCacheSubnetGroup(m map[string]any) error {
	g, err := AsElastiCacheSubnetGroup(m)
	if err != nil {
		return err
	}
	if len(g.Subnets) < 2 {
		return fmt.Errorf("subnets must contain at least two subnet IDs")
	}
	return nil
}

type ElastiCacheCluster struct {
	Engine         string   `json:"engine"`
	NodeType       string   `json:"node_type"`
	NumCacheNodes  int      `json:"num_cache_nodes"`
	SubnetGroup    string   `json:"subnet_group"`
	SecurityGroups []string `json:"security_groups"`
	Port           int      `json:"port,omitempty"`
}

func AsElastiCacheCluster(m map[string]any) (*ElastiCacheCluster, error) {
	return decode[ElastiCacheCluster](m)
}
func validateElastiCacheCluster(m map[string]any) error {
	c, err := AsElastiCacheCluster(m)
	if err != nil {
		return err
	}
	if c.Engine == "" || c.NodeType == "" || c.SubnetGroup == "" || len(c.SecurityGroups) == 0 {
		return fmt.Errorf("engine, node_type, subnet_group, and security_groups are required")
	}
	if c.NumCacheNodes < 1 {
		return fmt.Errorf("num_cache_nodes must be >= 1")
	}
	return nil
}

type LambdaFunction struct {
	FunctionName string `json:"function_name"`
	Runtime      string `json:"runtime"`
	Handler      string `json:"handler"`
	Role         string `json:"role,omitempty"`
	Timeout      int    `json:"timeout,omitempty"`
	MemorySize   int    `json:"memory_size,omitempty"`
}

func AsLambdaFunction(m map[string]any) (*LambdaFunction, error) { return decode[LambdaFunction](m) }
func validateLambdaFunction(m map[string]any) error {
	l, err := AsLambdaFunction(m)
	if err != nil {
		return err
	}
	if l.FunctionName == "" || l.Runtime == "" || l.Handler == "" {
		return fmt.Errorf("function_name, runtime, and handler are required")
	}
	return nil
}

type LambdaPermission struct {
	FunctionName string `json:"function_name"`
	Principal    string `json:"principal"`
	SourceARN    string `json:"source_arn,omitempty"`
}

func AsLambdaPermission(m map[string]any) (*LambdaPermission, error) {
	return decode[LambdaPermission](m)
}
func validateLambdaPermission(m map[string]any) error {
	p, err := AsLambdaPermission(m)
	if err != nil {
		return err
	}
	if p.FunctionName == "" || p.Principal == "" {
		return fmt.Errorf("function_name and principal are required")
	}
	return nil
}

type APIGatewayV2API struct {
	Name         string `json:"name"`
	ProtocolType string `json:"protocol_type"`
	Target       string `json:"target,omitempty"`
}

func AsAPIGatewayV2API(m map[string]any) (*APIGatewayV2API, error) { return decode[APIGatewayV2API](m) }
func validateAPIGatewayV2API(m map[string]any) error {
	a, err := AsAPIGatewayV2API(m)
	if err != nil {
		return err
	}
	if a.Name == "" || a.ProtocolType == "" {
		return fmt.Errorf("name and protocol_type are required")
	}
	return nil
}

type APIGatewayV2Integration struct {
	APIID                string `json:"api_id"`
	IntegrationType      string `json:"integration_type"`
	IntegrationURI       string `json:"integration_uri"`
	PayloadFormatVersion string `json:"payload_format_version,omitempty"`
}

func AsAPIGatewayV2Integration(m map[string]any) (*APIGatewayV2Integration, error) {
	return decode[APIGatewayV2Integration](m)
}
func validateAPIGatewayV2Integration(m map[string]any) error {
	i, err := AsAPIGatewayV2Integration(m)
	if err != nil {
		return err
	}
	if i.APIID == "" || i.IntegrationType == "" || i.IntegrationURI == "" {
		return fmt.Errorf("api_id, integration_type, and integration_uri are required")
	}
	return nil
}

type APIGatewayV2Route struct {
	APIID    string `json:"api_id"`
	RouteKey string `json:"route_key"`
	Target   string `json:"target"`
}

func AsAPIGatewayV2Route(m map[string]any) (*APIGatewayV2Route, error) {
	return decode[APIGatewayV2Route](m)
}
func validateAPIGatewayV2Route(m map[string]any) error {
	r, err := AsAPIGatewayV2Route(m)
	if err != nil {
		return err
	}
	if r.APIID == "" || r.RouteKey == "" || r.Target == "" {
		return fmt.Errorf("api_id, route_key, and target are required")
	}
	return nil
}

type DynamoDBTable struct {
	Name          string `json:"name"`
	HashKey       string `json:"hash_key"`
	BillingMode   string `json:"billing_mode"`
	AttributeType string `json:"attribute_type,omitempty"`
}

func AsDynamoDBTable(m map[string]any) (*DynamoDBTable, error) { return decode[DynamoDBTable](m) }
func validateDynamoDBTable(m map[string]any) error {
	d, err := AsDynamoDBTable(m)
	if err != nil {
		return err
	}
	if d.Name == "" || d.HashKey == "" || d.BillingMode == "" {
		return fmt.Errorf("name, hash_key, and billing_mode are required")
	}
	return nil
}

type S3Bucket struct {
	Bucket string `json:"bucket"`
}

func AsS3Bucket(m map[string]any) (*S3Bucket, error) { return decode[S3Bucket](m) }
func validateS3Bucket(m map[string]any) error {
	s, err := AsS3Bucket(m)
	if err != nil {
		return err
	}
	if s.Bucket == "" {
		return fmt.Errorf("bucket required")
	}
	return nil
}

type OpenSearchDomain struct {
	DomainName    string `json:"domain_name"`
	EngineVersion string `json:"engine_version"`
	InstanceType  string `json:"instance_type"`
	InstanceCount int    `json:"instance_count"`
}

func AsOpenSearchDomain(m map[string]any) (*OpenSearchDomain, error) {
	return decode[OpenSearchDomain](m)
}
func validateOpenSearchDomain(m map[string]any) error {
	o, err := AsOpenSearchDomain(m)
	if err != nil {
		return err
	}
	if o.DomainName == "" || o.EngineVersion == "" || o.InstanceType == "" {
		return fmt.Errorf("domain_name, engine_version, and instance_type are required")
	}
	if o.InstanceCount < 1 {
		return fmt.Errorf("instance_count must be >= 1")
	}
	return nil
}

type EC2Instance struct {
	AMI               string   `json:"ami"`
	InstanceType      string   `json:"instance_type"`
	Subnet            string   `json:"subnet"`
	SecurityGroups    []string `json:"security_groups"`
	AssociatePublicIP bool     `json:"associate_public_ip,omitempty"`
}

func AsEC2Instance(m map[string]any) (*EC2Instance, error) { return decode[EC2Instance](m) }
func validateEC2Instance(m map[string]any) error {
	e, err := AsEC2Instance(m)
	if err != nil {
		return err
	}
	if e.AMI == "" || e.InstanceType == "" || e.Subnet == "" || len(e.SecurityGroups) == 0 {
		return fmt.Errorf("ami, instance_type, subnet, and security_groups are required")
	}
	return nil
}

type CloudFrontDistribution struct {
	OriginDNS string   `json:"origin_dns"`
	Aliases   []string `json:"aliases,omitempty"`
}

func AsCloudFrontDist(m map[string]any) (*CloudFrontDistribution, error) {
	return decode[CloudFrontDistribution](m)
}
func validateCloudFrontDist(m map[string]any) error {
	c, err := AsCloudFrontDist(m)
	if err != nil {
		return err
	}
	if c.OriginDNS == "" {
		return fmt.Errorf("origin_dns required")
	}
	return nil
}

type IAMRole struct {
	Service           string   `json:"service"`
	ManagedPolicyArns []string `json:"managed_policy_arns,omitempty"`
}

func AsIAMRole(m map[string]any) (*IAMRole, error) { return decode[IAMRole](m) }
func validateIAMRole(m map[string]any) error {
	r, err := AsIAMRole(m)
	if err != nil {
		return err
	}
	if r.Service == "" {
		return fmt.Errorf("service required")
	}
	return nil
}

type CloudWatchLogGroup struct {
	Name            string `json:"name"`
	RetentionInDays int    `json:"retention_in_days,omitempty"`
}

func AsCloudWatchLogGroup(m map[string]any) (*CloudWatchLogGroup, error) {
	return decode[CloudWatchLogGroup](m)
}
func validateCloudWatchLogGroup(m map[string]any) error {
	l, err := AsCloudWatchLogGroup(m)
	if err != nil {
		return err
	}
	if l.Name == "" {
		return fmt.Errorf("name required")
	}
	return nil
}

func decode[T any](m map[string]any) (*T, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("decode props: %w", err)
	}
	return &out, nil
}
