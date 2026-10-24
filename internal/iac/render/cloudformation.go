package render

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/oche/diagram/internal/iac/graph"
	"github.com/oche/diagram/internal/iac/graph/props"
)

type cfnTemplate struct {
	AWSTemplateFormatVersion string                 `yaml:"AWSTemplateFormatVersion"`
	Description              string                 `yaml:"Description"`
	Parameters               map[string]any         `yaml:"Parameters,omitempty"`
	Resources                map[string]cfnResource `yaml:"Resources"`
	Outputs                  map[string]any         `yaml:"Outputs,omitempty"`
}

type cfnResource struct {
	Type       string         `yaml:"Type"`
	DependsOn  []string       `yaml:"DependsOn,omitempty"`
	Properties map[string]any `yaml:"Properties,omitempty"`
}

type ref string

func (r ref) MarshalYAML() (any, error) {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!Ref", Value: string(r)}, nil
}

type getAtt struct {
	Resource string
	Attr     string
}

func (g getAtt) MarshalYAML() (any, error) {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!GetAtt",
		Value: g.Resource + "." + g.Attr,
	}, nil
}

func RenderCloudFormation(g *graph.ResourceGraph) (string, error) {
	if err := graph.Validate(g); err != nil {
		return "", err
	}
	ordered, err := TopoSort(g)
	if err != nil {
		return "", err
	}

	tmpl := cfnTemplate{
		AWSTemplateFormatVersion: "2010-09-09",
		Description:              g.Metadata.Description,
		Parameters:               map[string]any{},
		Resources:                map[string]cfnResource{},
		Outputs:                  map[string]any{},
	}

	for _, p := range g.Parameters {
		if p.Name == "db_password" {
			continue
		}
		name := cfnLogicalName(p.Name)
		param := map[string]any{
			"Type":        cfnParamType(p.Type),
			"Description": p.Description,
		}
		if p.Default != nil {
			param["Default"] = p.Default
		}
		tmpl.Parameters[name] = param
	}

	for _, r := range ordered {
		res, err := buildCFNResource(r)
		if err != nil {
			return "", fmt.Errorf("cfn %s (%s): %w", r.ID, r.Kind, err)
		}
		for _, dep := range r.DependsOn {
			res.DependsOn = append(res.DependsOn, cfnLogicalName(dep))
		}
		tmpl.Resources[cfnLogicalName(r.ID)] = res
	}

	for _, o := range g.Outputs {
		tmpl.Outputs[cfnLogicalName(o.Name)] = map[string]any{
			"Description": o.Description,
			"Value":       getAtt{Resource: cfnLogicalName(o.From), Attr: cfnAttr(o.Attr)},
		}
	}

	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(tmpl); err != nil {
		return "", err
	}
	_ = enc.Close()
	return buf.String(), nil
}

func cfnLogicalName(s string) string {
	var b strings.Builder
	upper := true
	for _, c := range s {
		if c == '_' || c == '-' {
			upper = true
			continue
		}
		if upper && c >= 'a' && c <= 'z' {
			b.WriteRune(c - 32)
			upper = false
			continue
		}
		upper = false
		b.WriteRune(c)
	}
	return b.String()
}

func cfnParamType(t string) string {
	switch t {
	case "number":
		return "Number"
	default:
		return "String"
	}
}

func cfnAttr(a string) string {
	switch a {
	case "endpoint":
		return "Endpoint.Address"
	case "dns_name":
		return "DNSName"
	case "arn":
		return "Arn"
	case "api_endpoint":
		return "ApiEndpoint"
	default:
		return a
	}
}

func buildCFNResource(r graph.Resource) (cfnResource, error) {
	switch r.Kind {
	case graph.KindVPC:
		p, err := props.AsVPC(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::EC2::VPC",
			Properties: map[string]any{
				"CidrBlock":          p.CIDRBlock,
				"EnableDnsHostnames": p.EnableDNSHostnames,
				"Tags":               []map[string]string{{"Key": "Name", "Value": r.ID}},
			},
		}, nil
	case graph.KindSubnet:
		p, err := props.AsSubnet(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::EC2::Subnet",
			Properties: map[string]any{
				"VpcId":               ref(cfnLogicalName(p.VPC)),
				"CidrBlock":           p.CIDRBlock,
				"AvailabilityZone":    availabilityZoneRef(p.AvailabilityZone),
				"MapPublicIpOnLaunch": p.MapPublicIPOnLaunch,
				"Tags":                []map[string]string{{"Key": "Name", "Value": r.ID}},
			},
		}, nil
	case graph.KindSecurityGroup:
		p, err := props.AsSecurityGroup(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		var ingress []map[string]any
		for _, rule := range p.Ingress {
			ingress = append(ingress, map[string]any{
				"IpProtocol": rule.Protocol,
				"FromPort":   rule.FromPort,
				"ToPort":     rule.ToPort,
				"CidrIp":     rule.CIDR,
			})
		}
		for _, rule := range p.IngressFromSG {
			ingress = append(ingress, map[string]any{
				"IpProtocol":            rule.Protocol,
				"FromPort":              rule.FromPort,
				"ToPort":                rule.ToPort,
				"SourceSecurityGroupId": ref(cfnLogicalName(rule.SG)),
			})
		}
		return cfnResource{
			Type: "AWS::EC2::SecurityGroup",
			Properties: map[string]any{
				"GroupDescription":     r.ID,
				"VpcId":                ref(cfnLogicalName(p.VPC)),
				"SecurityGroupIngress": ingress,
				"SecurityGroupEgress": []map[string]any{{
					"IpProtocol": "-1",
					"CidrIp":     "0.0.0.0/0",
				}},
			},
		}, nil
	case graph.KindALB:
		p, err := props.AsALB(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::ElasticLoadBalancingV2::LoadBalancer",
			Properties: map[string]any{
				"Name":           r.ID,
				"Type":           "application",
				"Scheme":         "internet-facing",
				"Subnets":        refsToAny(p.Subnets),
				"SecurityGroups": refsToAny(p.SecurityGroups),
			},
		}, nil
	case graph.KindTargetGroup:
		p, err := props.AsTargetGroup(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::ElasticLoadBalancingV2::TargetGroup",
			Properties: map[string]any{
				"Name":            r.ID,
				"VpcId":           ref(cfnLogicalName(p.VPC)),
				"Port":            p.Port,
				"Protocol":        p.Protocol,
				"TargetType":      p.TargetType,
				"HealthCheckPath": "/",
			},
		}, nil
	case graph.KindListener:
		p, err := props.AsListener(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::ElasticLoadBalancingV2::Listener",
			Properties: map[string]any{
				"LoadBalancerArn": ref(cfnLogicalName(p.LoadBalancer)),
				"Port":            p.Port,
				"Protocol":        p.Protocol,
				"Certificates":    []map[string]any{{"CertificateArn": "arn:aws:acm:us-east-1:123456789012:certificate/replace-me"}},
				"DefaultActions": []map[string]any{{
					"Type":           "forward",
					"TargetGroupArn": ref(cfnLogicalName(p.TargetGroup)),
				}},
			},
		}, nil
	case graph.KindECSCluster:
		return cfnResource{
			Type: "AWS::ECS::Cluster",
			Properties: map[string]any{
				"ClusterName": r.ID,
			},
		}, nil
	case graph.KindECSTaskDefinition:
		p, err := props.AsECSTaskDefinition(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		containerName := p.ContainerName
		if containerName == "" {
			containerName = p.Family
		}
		return cfnResource{
			Type: "AWS::ECS::TaskDefinition",
			Properties: map[string]any{
				"Family":                  p.Family,
				"Cpu":                     p.CPU,
				"Memory":                  p.Memory,
				"NetworkMode":             "awsvpc",
				"RequiresCompatibilities": []string{"FARGATE"},
				"ExecutionRoleArn":        "arn:aws:iam::123456789012:role/ecsTaskExecutionRole",
				"ContainerDefinitions": []map[string]any{{
					"Name":  containerName,
					"Image": p.Image,
					"PortMappings": []map[string]any{{
						"ContainerPort": p.ContainerPort,
						"Protocol":      "tcp",
					}},
				}},
			},
		}, nil
	case graph.KindECSService:
		p, err := props.AsECSService(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		service := cfnResource{
			Type: "AWS::ECS::Service",
			Properties: map[string]any{
				"ServiceName":    r.ID,
				"Cluster":        ref(cfnLogicalName(p.Cluster)),
				"TaskDefinition": ref(cfnLogicalName(p.TaskDefinition)),
				"DesiredCount":   p.DesiredCount,
				"LaunchType":     p.LaunchType,
				"NetworkConfiguration": map[string]any{
					"AwsvpcConfiguration": map[string]any{
						"Subnets":        refsToAny(p.Subnets),
						"SecurityGroups": refsToAny(p.SecurityGroups),
						"AssignPublicIp": ternaryString(p.AssignPublicIP, "ENABLED", "DISABLED"),
					},
				},
			},
		}
		if p.TargetGroup != "" {
			service.Properties["LoadBalancers"] = []map[string]any{{
				"TargetGroupArn": ref(cfnLogicalName(p.TargetGroup)),
				"ContainerName":  p.ContainerName,
				"ContainerPort":  p.ContainerPort,
			}}
		}
		return service, nil
	case graph.KindRDSSubnetGroup:
		p, err := props.AsRDSSubnetGroup(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::RDS::DBSubnetGroup",
			Properties: map[string]any{
				"DBSubnetGroupDescription": r.ID,
				"SubnetIds":                refsToAny(p.Subnets),
			},
		}, nil
	case graph.KindRDSInstance:
		p, err := props.AsRDSInstance(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		res := cfnResource{
			Type: "AWS::RDS::DBInstance",
			Properties: map[string]any{
				"Engine":                   p.Engine,
				"DBInstanceClass":          p.InstanceClass,
				"AllocatedStorage":         fmt.Sprintf("%d", p.AllocatedStorage),
				"DBName":                   p.DBName,
				"DBSubnetGroupName":        ref(cfnLogicalName(p.DBSubnetGroup)),
				"VPCSecurityGroups":        refsToAny(p.VPCSecurityGroups),
				"MasterUsername":           "app",
				"ManageMasterUserPassword": true,
				"StorageEncrypted":         p.StorageEncrypted,
				"PubliclyAccessible":       false,
			},
		}
		if p.EngineVersion != "" {
			res.Properties["EngineVersion"] = p.EngineVersion
		}
		return res, nil
	case graph.KindElastiCacheSubnetGroup:
		p, err := props.AsElastiCacheSubnetGroup(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::ElastiCache::SubnetGroup",
			Properties: map[string]any{
				"Description": r.ID,
				"SubnetIds":   refsToAny(p.Subnets),
			},
		}, nil
	case graph.KindElastiCacheCluster:
		p, err := props.AsElastiCacheCluster(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::ElastiCache::CacheCluster",
			Properties: map[string]any{
				"ClusterName":          r.ID,
				"Engine":               p.Engine,
				"CacheNodeType":        p.NodeType,
				"NumCacheNodes":        p.NumCacheNodes,
				"VpcSecurityGroupIds":  refsToAny(p.SecurityGroups),
				"CacheSubnetGroupName": ref(cfnLogicalName(p.SubnetGroup)),
			},
		}, nil
	case graph.KindLambdaFunction:
		p, err := props.AsLambdaFunction(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		role := p.Role
		if role == "" {
			role = "arn:aws:iam::123456789012:role/lambda-basic-execution"
		}
		timeout := p.Timeout
		if timeout == 0 {
			timeout = 10
		}
		memory := p.MemorySize
		if memory == 0 {
			memory = 256
		}
		return cfnResource{
			Type: "AWS::Lambda::Function",
			Properties: map[string]any{
				"FunctionName": p.FunctionName,
				"Runtime":      p.Runtime,
				"Handler":      p.Handler,
				"Role":         role,
				"Timeout":      timeout,
				"MemorySize":   memory,
				"Code": map[string]any{
					"ZipFile": "exports.handler = async () => ({ statusCode: 200, body: 'ok' });",
				},
			},
		}, nil
	case graph.KindLambdaPermission:
		p, err := props.AsLambdaPermission(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		propsMap := map[string]any{
			"Action":       "lambda:InvokeFunction",
			"FunctionName": ref(cfnLogicalName(p.FunctionName)),
			"Principal":    p.Principal,
		}
		if p.SourceARN != "" {
			propsMap["SourceArn"] = p.SourceARN
		}
		return cfnResource{Type: "AWS::Lambda::Permission", Properties: propsMap}, nil
	case graph.KindAPIGatewayV2API:
		p, err := props.AsAPIGatewayV2API(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::ApiGatewayV2::Api",
			Properties: map[string]any{
				"Name":         p.Name,
				"ProtocolType": p.ProtocolType,
			},
		}, nil
	case graph.KindAPIGatewayV2Integration:
		p, err := props.AsAPIGatewayV2Integration(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		pfv := p.PayloadFormatVersion
		if pfv == "" {
			pfv = "2.0"
		}
		return cfnResource{
			Type: "AWS::ApiGatewayV2::Integration",
			Properties: map[string]any{
				"ApiId":                ref(cfnLogicalName(p.APIID)),
				"IntegrationType":      p.IntegrationType,
				"IntegrationUri":       p.IntegrationURI,
				"PayloadFormatVersion": pfv,
			},
		}, nil
	case graph.KindAPIGatewayV2Route:
		p, err := props.AsAPIGatewayV2Route(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::ApiGatewayV2::Route",
			Properties: map[string]any{
				"ApiId":    ref(cfnLogicalName(p.APIID)),
				"RouteKey": p.RouteKey,
				"Target":   p.Target,
			},
		}, nil
	case graph.KindDynamoDBTable:
		p, err := props.AsDynamoDBTable(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		attrType := p.AttributeType
		if attrType == "" {
			attrType = "S"
		}
		return cfnResource{
			Type: "AWS::DynamoDB::Table",
			Properties: map[string]any{
				"TableName":   p.Name,
				"BillingMode": p.BillingMode,
				"AttributeDefinitions": []map[string]any{{
					"AttributeName": p.HashKey,
					"AttributeType": attrType,
				}},
				"KeySchema": []map[string]any{{
					"AttributeName": p.HashKey,
					"KeyType":       "HASH",
				}},
			},
		}, nil
	case graph.KindS3Bucket:
		p, err := props.AsS3Bucket(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{Type: "AWS::S3::Bucket", Properties: map[string]any{"BucketName": p.Bucket}}, nil
	case graph.KindOpenSearchDomain:
		p, err := props.AsOpenSearchDomain(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::OpenSearchService::Domain",
			Properties: map[string]any{
				"DomainName":    p.DomainName,
				"EngineVersion": p.EngineVersion,
				"ClusterConfig": map[string]any{
					"InstanceType":  p.InstanceType,
					"InstanceCount": p.InstanceCount,
				},
				"EBSOptions": map[string]any{
					"EBSEnabled": true,
					"VolumeSize": 20,
				},
			},
		}, nil
	case graph.KindEC2Instance:
		p, err := props.AsEC2Instance(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		return cfnResource{
			Type: "AWS::EC2::Instance",
			Properties: map[string]any{
				"ImageId":          p.AMI,
				"InstanceType":     p.InstanceType,
				"SubnetId":         ref(cfnLogicalName(p.Subnet)),
				"SecurityGroupIds": refsToAny(p.SecurityGroups),
			},
		}, nil
	case graph.KindCloudWatchLogGroup:
		p, err := props.AsCloudWatchLogGroup(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		retention := p.RetentionInDays
		if retention == 0 {
			retention = 14
		}
		return cfnResource{
			Type: "AWS::Logs::LogGroup",
			Properties: map[string]any{
				"LogGroupName":    p.Name,
				"RetentionInDays": retention,
			},
		}, nil
	case graph.KindCloudFrontDist:
		p, err := props.AsCloudFrontDist(r.Props)
		if err != nil {
			return cfnResource{}, err
		}
		originID := cfnLogicalName(r.ID) + "Origin"
		return cfnResource{
			Type: "AWS::CloudFront::Distribution",
			Properties: map[string]any{
				"DistributionConfig": map[string]any{
					"Enabled": true,
					"Aliases": p.Aliases,
					"Origins": []map[string]any{{
						"DomainName": getAtt{Resource: cfnLogicalName(p.OriginDNS), Attr: "DNSName"},
						"Id":         originID,
						"CustomOriginConfig": map[string]any{
							"HTTPPort":             80,
							"HTTPSPort":            443,
							"OriginProtocolPolicy": "https-only",
							"OriginSSLProtocols":   []string{"TLSv1.2"},
						},
					}},
					"DefaultCacheBehavior": map[string]any{
						"TargetOriginId":       originID,
						"ViewerProtocolPolicy": "redirect-to-https",
						"AllowedMethods":       []string{"GET", "HEAD", "OPTIONS"},
						"CachedMethods":        []string{"GET", "HEAD"},
						"ForwardedValues": map[string]any{
							"QueryString": true,
							"Cookies": map[string]any{
								"Forward": "all",
							},
						},
					},
					"Restrictions": map[string]any{
						"GeoRestriction": map[string]any{
							"RestrictionType": "none",
						},
					},
					"ViewerCertificate": map[string]any{
						"CloudFrontDefaultCertificate": true,
					},
				},
			},
		}, nil
	default:
		return cfnResource{}, fmt.Errorf("no cfn builder for kind %s", r.Kind)
	}
}

func refsToAny(ids []string) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = ref(cfnLogicalName(id))
	}
	return out
}

func ternaryString(cond bool, left, right string) string {
	if cond {
		return left
	}
	return right
}

func availabilityZoneRef(az string) any {
	if len(az) > 0 {
		last := az[len(az)-1]
		if last >= 'a' && last <= 'z' {
			return map[string]any{
				"Fn::Select": []any{
					int(last - 'a'),
					map[string]any{"Fn::GetAZs": ""},
				},
			}
		}
	}
	return az
}
