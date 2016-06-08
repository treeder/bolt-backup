package backup

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

// StartBackups db is the bolt db to backup, s3bucketName is where to store backups, interval is the interval between backups
func Open(dbPath string, mode os.FileMode, options *bolt.Options, s3BucketName string, interval time.Duration) (*bolt.DB, error) {
	auth, err := aws.EnvAuth()
	if err != nil {
		return nil, err
	}

	conn := s3.New(auth, aws.USEast)
	bucket := conn.Bucket(s3BucketName)
	err = bucket.PutBucket(s3.Private)
	if err != nil {
		return nil, err
	}

	s3path := "/bolt.db"

	// check for existing
	err = recoverFromS3(bucket, s3path, dbPath)
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}

	go startBackups(db, bucket, s3path, interval)
	return db, nil
}

func startBackups(db *bolt.DB, bucket *s3.Bucket, path string, interval time.Duration) {
	c := time.Tick(interval)
	for _ = range c {
		// todo: time this function, if it's longer than the duration log it so user knows it's taking too long
		err := backupDbToS3(db, bucket, path)
		if err != nil {
			log.Println("ERROR backing up:", err)
		}
	}
}

func backupDbToS3(db *bolt.DB, bucket *s3.Bucket, path string) error {
	log.Println("Starting backup.")
	r, w := io.Pipe()

	// start transaction
	tx, err := db.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	size := tx.Size()

	// write snapshot to pipe
	go func() {
		defer w.Close()
		_, err := tx.WriteTo(w)
		if err != nil {
			log.Println("Erroring writing to pipe", err)
		}
	}()

	// write to s3. This should block until above func exits
	err = bucket.PutReader(path, r, size, "application/octet-stream", s3.Private)
	if err != nil {
		return err
	}
	log.Println("Backup complete!")
	return nil
}

func recoverFromS3(bucket *s3.Bucket, s3path, dbPath string) error {
	rc, err := bucket.GetReader(s3path)
	if err != nil {
		log.Println("DB not found in s3, nothing to restore.", err)
		return nil
	}
	defer rc.Close()
	out, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	defer out.Close()
	x, err := io.Copy(out, rc)
	if err != nil {
		return err
	}
	log.Println("Recovered", x, "bytes of bolt")
	return nil
}
