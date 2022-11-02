package lyvecloud

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestRetryWhenAWSErrCodeEquals(t *testing.T) { // nosemgrep:aws-in-func-name
	var retryCount int32

	testCases := []struct {
		Name        string
		F           func() (interface{}, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func() (interface{}, error) {
				return nil, nil
			},
		},
		{
			Name: "non-retryable other error",
			F: func() (interface{}, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "non-retryable AWS error",
			F: func() (interface{}, error) {
				return nil, awserr.New("Testing", "Testing", nil)
			},
			ExpectError: true,
		},
		{
			Name: "retryable AWS error timeout",
			F: func() (interface{}, error) {
				return nil, awserr.New("TestCode1", "TestMessage", nil)
			},
			ExpectError: true,
		},
		{
			Name: "retryable AWS error success",
			F: func() (interface{}, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return nil, awserr.New("TestCode2", "TestMessage", nil)
				}

				return nil, nil
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := RetryWhenAWSErrCodeEquals(5*time.Second, testCase.F, "TestCode1", "TestCode2")

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestRetryWhenAWSErrMessageContains(t *testing.T) { // nosemgrep:aws-in-func-name
	var retryCount int32

	testCases := []struct {
		Name        string
		F           func() (interface{}, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func() (interface{}, error) {
				return nil, nil
			},
		},
		{
			Name: "non-retryable other error",
			F: func() (interface{}, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "non-retryable AWS error",
			F: func() (interface{}, error) {
				return nil, awserr.New("TestCode1", "Testing", nil)
			},
			ExpectError: true,
		},
		{
			Name: "retryable AWS error timeout",
			F: func() (interface{}, error) {
				return nil, awserr.New("TestCode1", "TestMessage1", nil)
			},
			ExpectError: true,
		},
		{
			Name: "retryable AWS error success",
			F: func() (interface{}, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return nil, awserr.New("TestCode1", "TestMessage1", nil)
				}

				return nil, nil
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := RetryWhenAWSErrMessageContains(5*time.Second, testCase.F, "TestCode1", "TestMessage1")

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestRetryUntilNotFound(t *testing.T) {
	var retryCount int32

	testCases := []struct {
		Name        string
		F           func() (interface{}, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func() (interface{}, error) {
				return nil, nil
			},
			ExpectError: true,
		},
		{
			Name: "other error",
			F: func() (interface{}, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "AWS error",
			F: func() (interface{}, error) {
				return nil, awserr.New("Testing", "Testing", nil)
			},
			ExpectError: true,
		},
		{
			Name: "NotFoundError",
			F: func() (interface{}, error) {
				return nil, &resource.NotFoundError{}
			},
		},
		{
			Name: "retryable NotFoundError",
			F: func() (interface{}, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return nil, nil
				}

				return nil, &resource.NotFoundError{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := RetryUntilNotFound(5*time.Second, testCase.F)

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestRetryWhenNotFound(t *testing.T) {
	var retryCount int32

	testCases := []struct {
		Name        string
		F           func() (interface{}, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func() (interface{}, error) {
				return nil, nil
			},
		},
		{
			Name: "non-retryable other error",
			F: func() (interface{}, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "non-retryable AWS error",
			F: func() (interface{}, error) {
				return nil, awserr.New("Testing", "Testing", nil)
			},
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError timeout",
			F: func() (interface{}, error) {
				return nil, &resource.NotFoundError{}
			},
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError success",
			F: func() (interface{}, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return nil, &resource.NotFoundError{}
				}

				return nil, nil
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := RetryWhenNotFound(5*time.Second, testCase.F)

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestRetryWhenNewResourceNotFound(t *testing.T) {
	var retryCount int32

	testCases := []struct {
		Name        string
		F           func() (interface{}, error)
		NewResource bool
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func() (interface{}, error) {
				return nil, nil
			},
		},
		{
			Name: "no error new resource",
			F: func() (interface{}, error) {
				return nil, nil
			},
			NewResource: true,
		},
		{
			Name: "non-retryable other error",
			F: func() (interface{}, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "non-retryable other error new resource",
			F: func() (interface{}, error) {
				return nil, errors.New("TestCode")
			},
			NewResource: true,
			ExpectError: true,
		},
		{
			Name: "non-retryable AWS error",
			F: func() (interface{}, error) {
				return nil, awserr.New("Testing", "Testing", nil)
			},
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError not new resource",
			F: func() (interface{}, error) {
				return nil, &resource.NotFoundError{}
			},
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError new resource timeout",
			F: func() (interface{}, error) {
				return nil, &resource.NotFoundError{}
			},
			NewResource: true,
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError success new resource",
			F: func() (interface{}, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return nil, &resource.NotFoundError{}
				}

				return nil, nil
			},
			NewResource: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := RetryWhenNotFound(5*time.Second, testCase.F)

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestRetryConfigContext_error(t *testing.T) {
	t.Parallel()

	expected := fmt.Errorf("nope")
	f := func() *resource.RetryError {
		return resource.NonRetryableError(expected)
	}

	errCh := make(chan error)
	go func() {
		errCh <- RetryConfigContext(context.Background(), 0*time.Second, 0*time.Second, 0*time.Second, 0*time.Second, 1*time.Second, f)
	}()

	select {
	case err := <-errCh:
		if err != expected {
			t.Fatalf("bad: %#v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func TestNotFound(t *testing.T) {
	testCases := []struct {
		Name     string
		Err      error
		Expected bool
	}{
		{
			Name: "nil error",
			Err:  nil,
		},
		{
			Name: "other error",
			Err:  errors.New("test"),
		},
		{
			Name:     "not found error",
			Err:      &resource.NotFoundError{LastError: errors.New("test")},
			Expected: true,
		},
		{
			Name: "wrapped other error",
			Err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			Name:     "wrapped not found error",
			Err:      fmt.Errorf("test: %w", &resource.NotFoundError{LastError: errors.New("test")}),
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := NotFound(testCase.Err)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestTimedOut(t *testing.T) {
	testCases := []struct {
		Name     string
		Err      error
		Expected bool
	}{
		{
			Name: "nil error",
			Err:  nil,
		},
		{
			Name: "other error",
			Err:  errors.New("test"),
		},
		{
			Name:     "timeout error",
			Err:      &resource.TimeoutError{},
			Expected: true,
		},
		{
			Name: "timeout error non-nil last error",
			Err:  &resource.TimeoutError{LastError: errors.New("test")},
		},
		{
			Name: "wrapped other error",
			Err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			Name: "wrapped timeout error",
			Err:  fmt.Errorf("test: %w", &resource.TimeoutError{}),
		},
		{
			Name: "wrapped timeout error non-nil last error",
			Err:  fmt.Errorf("test: %w", &resource.TimeoutError{LastError: errors.New("test")}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := TimedOut(testCase.Err)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
