package function
/* Project: hnsdbc */

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"

	"golang.org/x/crypto/bcrypt"
)

const projectID = "hnsdbc"

type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

/* collection<users>/doc<user.id>
 * Id use email
 */
type User struct {
	Id     string    `firestore:"-" json:"-"`
	Name   string    `firestore:"name" json:"name"`
	Email  string    `firestore:"email" json:"email"`
	Pass   string    `firestore:"pass" json:"pass"`
	Role   int       `firestore:"role" json:"role,string"`
	Date   time.Time `firestore:"date" json:"date"`
	Update time.Time `firestore:"update,serverTimestamp" json:"update"`
}

// password encryption
func hashAndSalt(pwd string) (string, error) {
	// MinCost (4) DefaultCost(10) MaxCost(14)
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

/* Only handles objects cantaining 'users' name
 */
func saveUsersFromCsv(ctx context.Context, obj *storage.ObjectHandle) error {
	r, err := obj.NewReader(ctx)
	//attrs, err := obj.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("Storage reader is error: %v", err)
	}
	defer r.Close()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("New firestore is error: %v", err)
	}
	defer client.Close()

	rd := csv.NewReader(io.Reader(r))
	batch := client.Batch()

	for {
		record, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Csv fila on reading is error: %v", err)
		}
		u := &User{}
		u.Email = record[0]
		hash, _ := hashAndSalt(record[1])
		u.Pass = hash
		role, _ := strconv.Atoi(record[2])
		u.Role = role
		u.Name = strings.Split(u.Email, "@")[0]
		u.Date = time.Now()

		uRef := client.Collection("users").Doc(record[0])
		batch.Set(uRef, u)
	}
	_, err = batch.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Batch commit is error: %v", err)
	}
	return nil
}

func UploadUsers(ctx context.Context, e *GCSEvent) error {
	log.Println("Start upload users from storage")
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	src := client.Bucket(e.Bucket).Object(e.Name)
	if err = saveUsersFromCsv(ctx, src); err != nil {
		return err
	}
	log.Println("End upload users from storage")

	return nil
}

func DeleteUsers(ctx context.Context, e *GCSEvent) error {

	log.Println("Start delete users from firestore")
	clStorage, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer clStorage.Close()

	obj := clStorage.Bucket(e.Bucket).Object(e.Name)

	r, _ := obj.NewReader(ctx)
	//attrs, err := obj.Attrs(ctx)
	defer r.Close()

	clFirestore, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer clFirestore.Close()

	rd := csv.NewReader(io.Reader(r))
	for {
		record, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		id := record[0]
		_, err = clFirestore.Collection("users").Doc(id).Delete(ctx)
		if err != nil {
			return err
		}
	}
	log.Println("Start delete users from firestore")
	return nil
}
