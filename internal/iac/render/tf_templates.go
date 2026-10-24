package render

import "github.com/oche/diagram/internal/iac/graph"

func init() {
	tfTemplates[graph.KindVPC] = mustTemplate("vpc", `resource "aws_vpc" "{{.R.ID}}" {
  cidr_block           = {{.P.CIDRBlock | hcl}}
  enable_dns_hostnames = {{if .P.EnableDNSHostnames}}true{{else}}false{{end}}
  tags = {
    Name = {{.R.ID | hcl}}
  }
}
`)

	tfTemplates[graph.KindSubnet] = mustTemplate("subnet", `resource "aws_subnet" "{{.R.ID}}" {
  vpc_id                  = aws_vpc.{{.P.VPC}}.id
  cidr_block              = {{.P.CIDRBlock | hcl}}
  availability_zone       = {{.P.AvailabilityZone | hcl}}
  map_public_ip_on_launch = {{if .P.MapPublicIPOnLaunch}}true{{else}}false{{end}}
  tags = {
    Name = {{.R.ID | hcl}}
  }
}
`)

	tfTemplates[graph.KindSecurityGroup] = mustTemplate("sg", `resource "aws_security_group" "{{.R.ID}}" {
  name   = {{.R.ID | hcl}}
  vpc_id = aws_vpc.{{.P.VPC}}.id
{{range .P.Ingress}}
  ingress {
    from_port   = {{.FromPort}}
    to_port     = {{.ToPort}}
    protocol    = {{.Protocol | hcl}}
    cidr_blocks = [{{.CIDR | hcl}}]
  }
{{end}}
{{range .P.IngressFromSG}}
  ingress {
    from_port       = {{.FromPort}}
    to_port         = {{.ToPort}}
    protocol        = {{.Protocol | hcl}}
    security_groups = [aws_security_group.{{.SG}}.id]
  }
{{end}}
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`)

	tfTemplates[graph.KindALB] = mustTemplate("alb", `resource "aws_lb" "{{.R.ID}}" {
  name               = {{.R.ID | hcl}}
  internal           = false
  load_balancer_type = "application"
  subnets            = {{tfSubnets .P.Subnets}}
  security_groups    = {{tfSGs .P.SecurityGroups}}
}
`)

	tfTemplates[graph.KindTargetGroup] = mustTemplate("tg", `resource "aws_lb_target_group" "{{.R.ID}}" {
  name        = {{.R.ID | hcl}}
  port        = {{.P.Port}}
  protocol    = {{.P.Protocol | hcl}}
  target_type = {{.P.TargetType | hcl}}
  vpc_id      = aws_vpc.{{.P.VPC}}.id

  health_check {
    path = "/"
  }
}
`)

	tfTemplates[graph.KindListener] = mustTemplate("listener", `resource "aws_lb_listener" "{{.R.ID}}" {
  load_balancer_arn = aws_lb.{{.P.LoadBalancer}}.arn
  port              = {{.P.Port}}
  protocol          = {{.P.Protocol | hcl}}
  certificate_arn   = "arn:aws:acm:us-east-1:123456789012:certificate/replace-me"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.{{.P.TargetGroup}}.arn
  }
}
`)

	tfTemplates[graph.KindECSCluster] = mustTemplate("cluster", `resource "aws_ecs_cluster" "{{.R.ID}}" {
  name = {{.R.ID | hcl}}
}
`)

	tfTemplates[graph.KindECSTaskDefinition] = mustTemplate("task", `resource "aws_ecs_task_definition" "{{.R.ID}}" {
  family                   = {{.P.Family | hcl}}
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = {{.P.CPU | hcl}}
  memory                   = {{.P.Memory | hcl}}
  execution_role_arn       = "arn:aws:iam::123456789012:role/ecsTaskExecutionRole"
  container_definitions    = jsonencode([
    {
      name         = {{if .P.ContainerName}}{{.P.ContainerName | hcl}}{{else}}{{.P.Family | hcl}}{{end}}
      image        = {{.P.Image | hcl}}
      essential    = true
      portMappings = [
        {
          containerPort = {{.P.ContainerPort}}
          hostPort      = {{.P.ContainerPort}}
          protocol      = "tcp"
        }
      ]
    }
  ])
}
`)

	tfTemplates[graph.KindECSService] = mustTemplate("service", `resource "aws_ecs_service" "{{.R.ID}}" {
  name            = {{.R.ID | hcl}}
  cluster         = aws_ecs_cluster.{{.P.Cluster}}.id
  task_definition = aws_ecs_task_definition.{{.P.TaskDefinition}}.arn
  desired_count   = {{.P.DesiredCount}}
  launch_type     = {{.P.LaunchType | hcl}}

  network_configuration {
    subnets          = {{tfSubnets .P.Subnets}}
    security_groups  = {{tfSGs .P.SecurityGroups}}
    assign_public_ip = {{if .P.AssignPublicIP}}true{{else}}false{{end}}
  }
{{if .P.TargetGroup}}
  load_balancer {
    target_group_arn = aws_lb_target_group.{{.P.TargetGroup}}.arn
    container_name   = {{.P.ContainerName | hcl}}
    container_port   = {{.P.ContainerPort}}
  }
{{end}}
}
`)

	tfTemplates[graph.KindRDSSubnetGroup] = mustTemplate("rds_subnet_group", `resource "aws_db_subnet_group" "{{.R.ID}}" {
  name       = {{.R.ID | hcl}}
  subnet_ids = {{tfSubnets .P.Subnets}}
}
`)

	tfTemplates[graph.KindRDSInstance] = mustTemplate("rds_instance", `resource "aws_db_instance" "{{.R.ID}}" {
  identifier             = {{.R.ID | hcl}}
  engine                 = {{.P.Engine | hcl}}
{{if .P.EngineVersion}}
  engine_version         = {{.P.EngineVersion | hcl}}
{{end}}
  instance_class         = {{.P.InstanceClass | hcl}}
  allocated_storage      = {{.P.AllocatedStorage}}
  db_name                = {{.P.DBName | hcl}}
  db_subnet_group_name   = aws_db_subnet_group.{{.P.DBSubnetGroup}}.name
  vpc_security_group_ids = {{tfSGs .P.VPCSecurityGroups}}
  skip_final_snapshot    = {{if .P.SkipFinalSnapshot}}true{{else}}false{{end}}
  storage_encrypted      = {{if .P.StorageEncrypted}}true{{else}}false{{end}}
  username               = "app"
  password               = var.db_password
}
`)

	tfTemplates[graph.KindElastiCacheSubnetGroup] = mustTemplate("elasticache_subnet_group", `resource "aws_elasticache_subnet_group" "{{.R.ID}}" {
  name       = {{.R.ID | hcl}}
  subnet_ids = {{tfSubnets .P.Subnets}}
}
`)

	tfTemplates[graph.KindElastiCacheCluster] = mustTemplate("elasticache_cluster", `resource "aws_elasticache_cluster" "{{.R.ID}}" {
  cluster_id           = {{.R.ID | hcl}}
  engine               = {{.P.Engine | hcl}}
  node_type            = {{.P.NodeType | hcl}}
  num_cache_nodes      = {{.P.NumCacheNodes}}
  port                 = {{.P.Port}}
  subnet_group_name    = aws_elasticache_subnet_group.{{.P.SubnetGroup}}.name
  security_group_ids   = {{tfSGs .P.SecurityGroups}}
}
`)

	tfTemplates[graph.KindLambdaFunction] = mustTemplate("lambda_function", `resource "aws_lambda_function" "{{.R.ID}}" {
  function_name = {{.P.FunctionName | hcl}}
  role          = {{if .P.Role}}{{.P.Role | hcl}}{{else}}"arn:aws:iam::123456789012:role/lambda-basic-execution"{{end}}
  runtime       = {{.P.Runtime | hcl}}
  handler       = {{.P.Handler | hcl}}
  filename      = "lambda.zip"
  source_code_hash = filebase64sha256("lambda.zip")
  timeout       = {{if .P.Timeout}}{{.P.Timeout}}{{else}}10{{end}}
  memory_size   = {{if .P.MemorySize}}{{.P.MemorySize}}{{else}}256{{end}}
}
`)

	tfTemplates[graph.KindLambdaPermission] = mustTemplate("lambda_permission", `resource "aws_lambda_permission" "{{.R.ID}}" {
  statement_id  = {{printf "%s-allow" .R.ID | hcl}}
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.{{.P.FunctionName}}.function_name
  principal     = {{.P.Principal | hcl}}
{{if .P.SourceARN}}
  source_arn    = {{.P.SourceARN | hcl}}
{{end}}
}
`)

	tfTemplates[graph.KindAPIGatewayV2API] = mustTemplate("apigwv2_api", `resource "aws_apigatewayv2_api" "{{.R.ID}}" {
  name          = {{.P.Name | hcl}}
  protocol_type = {{.P.ProtocolType | hcl}}
}
`)

	tfTemplates[graph.KindAPIGatewayV2Integration] = mustTemplate("apigwv2_integration", `resource "aws_apigatewayv2_integration" "{{.R.ID}}" {
  api_id                 = aws_apigatewayv2_api.{{.P.APIID}}.id
  integration_type       = {{.P.IntegrationType | hcl}}
  integration_uri        = {{.P.IntegrationURI | hcl}}
  payload_format_version = {{if .P.PayloadFormatVersion}}{{.P.PayloadFormatVersion | hcl}}{{else}}"2.0"{{end}}
}
`)

	tfTemplates[graph.KindAPIGatewayV2Route] = mustTemplate("apigwv2_route", `resource "aws_apigatewayv2_route" "{{.R.ID}}" {
  api_id    = aws_apigatewayv2_api.{{.P.APIID}}.id
  route_key = {{.P.RouteKey | hcl}}
  target    = {{.P.Target | hcl}}
}
`)

	tfTemplates[graph.KindDynamoDBTable] = mustTemplate("dynamodb_table", `resource "aws_dynamodb_table" "{{.R.ID}}" {
  name         = {{.P.Name | hcl}}
  billing_mode = {{.P.BillingMode | hcl}}
  hash_key     = {{.P.HashKey | hcl}}

  attribute {
    name = {{.P.HashKey | hcl}}
    type = {{if .P.AttributeType}}{{.P.AttributeType | hcl}}{{else}}"S"{{end}}
  }
}
`)

	tfTemplates[graph.KindS3Bucket] = mustTemplate("s3_bucket", `resource "aws_s3_bucket" "{{.R.ID}}" {
  bucket = {{.P.Bucket | hcl}}
}
`)

	tfTemplates[graph.KindOpenSearchDomain] = mustTemplate("opensearch_domain", `resource "aws_opensearch_domain" "{{.R.ID}}" {
  domain_name    = {{.P.DomainName | hcl}}
  engine_version = {{.P.EngineVersion | hcl}}

  cluster_config {
    instance_type  = {{.P.InstanceType | hcl}}
    instance_count = {{.P.InstanceCount}}
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 20
  }
}
`)

	tfTemplates[graph.KindEC2Instance] = mustTemplate("ec2_instance", `resource "aws_instance" "{{.R.ID}}" {
  ami                         = {{.P.AMI | hcl}}
  instance_type               = {{.P.InstanceType | hcl}}
  subnet_id                   = aws_subnet.{{.P.Subnet}}.id
  vpc_security_group_ids      = {{tfSGs .P.SecurityGroups}}
  associate_public_ip_address = {{if .P.AssociatePublicIP}}true{{else}}false{{end}}
}
`)

	tfTemplates[graph.KindCloudWatchLogGroup] = mustTemplate("log_group", `resource "aws_cloudwatch_log_group" "{{.R.ID}}" {
  name              = {{.P.Name | hcl}}
  retention_in_days = {{if .P.RetentionInDays}}{{.P.RetentionInDays}}{{else}}14{{end}}
}
`)

	tfTemplates[graph.KindCloudFrontDist] = mustTemplate("cloudfront", `resource "aws_cloudfront_distribution" "{{.R.ID}}" {
  enabled = true
  aliases = {{tfAliases .P.Aliases}}

  origin {
    domain_name = aws_lb.{{.P.OriginDNS}}.dns_name
    origin_id   = {{printf "%s-origin" .R.ID | hcl}}

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = {{printf "%s-origin" .R.ID | hcl}}
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = true
      cookies {
        forward = "all"
      }
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`)
}
