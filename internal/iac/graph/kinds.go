package graph

type ResourceKind string

const (
	KindVPC                     ResourceKind = "aws_vpc"
	KindSubnet                  ResourceKind = "aws_subnet"
	KindInternetGateway         ResourceKind = "aws_internet_gateway"
	KindNATGateway              ResourceKind = "aws_nat_gateway"
	KindRouteTable              ResourceKind = "aws_route_table"
	KindSecurityGroup           ResourceKind = "aws_security_group"
	KindALB                     ResourceKind = "aws_alb"
	KindTargetGroup             ResourceKind = "aws_target_group"
	KindListener                ResourceKind = "aws_listener"
	KindECSCluster              ResourceKind = "aws_ecs_cluster"
	KindECSService              ResourceKind = "aws_ecs_service"
	KindECSTaskDefinition       ResourceKind = "aws_ecs_task_definition"
	KindRDSInstance             ResourceKind = "aws_rds_instance"
	KindRDSSubnetGroup          ResourceKind = "aws_rds_subnet_group"
	KindElastiCacheCluster      ResourceKind = "aws_elasticache_cluster"
	KindElastiCacheSubnetGroup  ResourceKind = "aws_elasticache_subnet_group"
	KindLambdaFunction          ResourceKind = "aws_lambda_function"
	KindLambdaPermission        ResourceKind = "aws_lambda_permission"
	KindAPIGatewayV2API         ResourceKind = "aws_apigatewayv2_api"
	KindAPIGatewayV2Route       ResourceKind = "aws_apigatewayv2_route"
	KindAPIGatewayV2Integration ResourceKind = "aws_apigatewayv2_integration"
	KindDynamoDBTable           ResourceKind = "aws_dynamodb_table"
	KindS3Bucket                ResourceKind = "aws_s3_bucket"
	KindOpenSearchDomain        ResourceKind = "aws_opensearch_domain"
	KindEC2Instance             ResourceKind = "aws_ec2_instance"
	KindCloudFrontDist          ResourceKind = "aws_cloudfront_distribution"
	KindIAMRole                 ResourceKind = "aws_iam_role"
	KindCloudWatchLogGroup      ResourceKind = "aws_cloudwatch_log_group"
)

func Known() map[ResourceKind]bool {
	return map[ResourceKind]bool{
		KindVPC:                     true,
		KindSubnet:                  true,
		KindInternetGateway:         true,
		KindNATGateway:              true,
		KindRouteTable:              true,
		KindSecurityGroup:           true,
		KindALB:                     true,
		KindTargetGroup:             true,
		KindListener:                true,
		KindECSCluster:              true,
		KindECSService:              true,
		KindECSTaskDefinition:       true,
		KindRDSInstance:             true,
		KindRDSSubnetGroup:          true,
		KindElastiCacheCluster:      true,
		KindElastiCacheSubnetGroup:  true,
		KindLambdaFunction:          true,
		KindLambdaPermission:        true,
		KindAPIGatewayV2API:         true,
		KindAPIGatewayV2Route:       true,
		KindAPIGatewayV2Integration: true,
		KindDynamoDBTable:           true,
		KindS3Bucket:                true,
		KindOpenSearchDomain:        true,
		KindEC2Instance:             true,
		KindCloudFrontDist:          true,
		KindIAMRole:                 true,
		KindCloudWatchLogGroup:      true,
	}
}
