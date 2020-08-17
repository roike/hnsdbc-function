package function

/* project: hnsdbc */

import (
	"context"
	"path"
	"testing"
)

func TestUploadUsers(t *testing.T) {
	t.Logf("Upload users from storage.")
	ctx := context.Background()
	e := &GCSEvent{}
	e.Bucket = "hnsdbc.appspot.com"
	e.Name = path.Join("test", "test_users.csv")
	err := UploadUsers(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteUsers(t *testing.T) {
	t.Logf("Delete users from firestore.")
	ctx := context.Background()
	e := &GCSEvent{}
	e.Bucket = "hnsdbc.appspot.com"
	e.Name = path.Join("test", "test_users.csv")
	err := DeleteUsers(ctx, e)
	if err != nil {
		t.Fatal(err)
	}

}
