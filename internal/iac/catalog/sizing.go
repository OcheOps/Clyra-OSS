package catalog

type Size string

const (
	SizeSmall  Size = "small"
	SizeMedium Size = "medium"
	SizeLarge  Size = "large"
)

type Instances struct {
	RDS string
	ECS string
	EC2 string
}

func Resolve(s Size) Instances {
	switch s {
	case SizeLarge:
		return Instances{RDS: "db.r6g.large", ECS: "1024/2048", EC2: "m6g.large"}
	case SizeMedium:
		return Instances{RDS: "db.t4g.small", ECS: "512/1024", EC2: "t3.medium"}
	default:
		return Instances{RDS: "db.t4g.micro", ECS: "256/512", EC2: "t3.small"}
	}
}
