package lyvecloud

import (
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

type KeyValueTags map[string]*TagData

type TagData struct {
	// Additional boolean field names and values associated with this tag.
	// Each service is responsible for properly handling this data.
	AdditionalBoolFields map[string]*bool

	// Additional string field names and values associated with this tag.
	// Each service is responsible for properly handling this data.
	AdditionalStringFields map[string]*string

	// Tag value.
	Value *string
}

// BucketUpdateTags updates S3 bucket tags.
// The identifier is the bucket name.
func BucketUpdateTags(conn *s3.S3, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if len(newTags) > 0 {
		input := &s3.PutBucketTaggingInput{
			Bucket: aws.String(identifier),
			Tagging: &s3.Tagging{
				TagSet: Tags(newTags),
			},
		}

		_, err := conn.PutBucketTagging(input)

		if err != nil {
			return fmt.Errorf("error setting resource tags (%s): %w", identifier, err)
		}
	} else if len(oldTags) > 0 {
		input := &s3.DeleteBucketTaggingInput{
			Bucket: aws.String(identifier),
		}

		_, err := conn.DeleteBucketTagging(input)

		if err != nil {
			return fmt.Errorf("error deleting resource tags (%s): %w", identifier, err)
		}
	}

	return nil
}

func New(i interface{}) KeyValueTags {
	switch value := i.(type) {
	case map[string]*string:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			strPtr := v

			if strPtr == nil {
				kvtm[k] = nil
				continue
			}

			kvtm[k] = &TagData{Value: strPtr}
		}

		return kvtm
	case map[string]interface{}:
		kvtm := make(KeyValueTags, len(value))

		for k, v := range value {
			kvtm[k] = &TagData{}

			str, ok := v.(string)

			if ok {
				kvtm[k].Value = &str
			}
		}

		return kvtm
	default:
		return make(KeyValueTags)
	}
}

// Tags returns s3 service tags.
func Tags(tags KeyValueTags) []*s3.Tag {
	result := make([]*s3.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// Map returns tag keys mapped to their values.
func (tags KeyValueTags) Map() map[string]string {
	result := make(map[string]string, len(tags))

	for k, v := range tags {
		if v == nil || v.Value == nil {
			result[k] = ""
			continue
		}

		result[k] = *v.Value
	}

	return result
}

// URLEncode returns the KeyValueTags encoded as URL Query parameters.
func (tags KeyValueTags) URLEncode() string {
	values := url.Values{}

	for k, v := range tags {
		if v == nil || v.Value == nil {
			continue
		}

		values.Add(k, *v.Value)
	}

	return values.Encode()
}

// ObjectListTags lists S3 object tags.
func ObjectListTags(conn *s3.S3, bucket, key string) (KeyValueTags, error) {
	input := &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	var output *s3.GetObjectTaggingOutput

	output, err := conn.GetObjectTagging(input)

	if tfawserr.ErrCodeEquals(err, "NoSuchTagSet") {
		return New(nil), nil
	}

	if err != nil {
		return New(nil), err
	}
	return KeyValueBucketTags(output.TagSet), nil
}

// KeyValueTags creates KeyValueTags from s3 service tags.
func KeyValueBucketTags(tags []*s3.Tag) KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return New(m)
}

// ObjectUpdateTags updates S3 object tags.
func ObjectUpdateTags(conn *s3.S3, bucket, key string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	if len(newTags) > 0 {
		input := &s3.PutObjectTaggingInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Tagging: &s3.Tagging{
				TagSet: Tags(newTags),
			},
		}

		_, err := conn.PutObjectTagging(input)

		if err != nil {
			return fmt.Errorf("error setting resource tags (%s/%s): %w", bucket, key, err)
		}
	} else if len(oldTags) > 0 {
		input := &s3.DeleteObjectTaggingInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}

		_, err := conn.DeleteObjectTagging(input)

		if err != nil {
			return fmt.Errorf("error deleting resource tags (%s/%s): %w", bucket, key, err)
		}
	}

	return nil
}
