package lyvecloud

import "github.com/aws/aws-sdk-go/aws"

// CheckCredentials checks that the client being used for the calling resource is nil, which is caused by missing credentials.
func CheckCredentials(cType string, client Client) bool {
	if cType == "s3" {
		if client.S3Client == nil {
			return true
		} else {
			return false
		}
	} else if cType == "acc" {
		if client.AccAPIV1Client == nil {
			return true
		} else {
			return false
		}
	} else if cType == "accv2" {
		if client.AccAPIV2Client == nil {
			return true
		} else {
			return false
		}
	}

	return true
}

// Expands a map of string to interface to a map of string to *string
func ExpandStringMap(m map[string]interface{}) map[string]*string {
	stringMap := make(map[string]*string, len(m))
	for k, v := range m {
		stringMap[k] = aws.String(v.(string))
	}
	return stringMap
}

func PointersMapToStringList(pointers map[string]*string) map[string]interface{} {
	list := make(map[string]interface{}, len(pointers))
	for i, v := range pointers {
		list[i] = *v
	}
	return list
}
