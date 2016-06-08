package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/treeder/bolt-backup"
)

type User struct {
	ID   int    `json:"id"`
	Name string `form:"name" json:"name" binding:"required"`
}

func main() {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	dbName := "my.db"

	// Start auto backup
	backupBucketName := os.Getenv("AWS_BUCKET_NAME")
	if backupBucketName == "" {
		log.Fatalln("AWS_BUCKET_NAME must be specified")
	}
	db, err := backup.Open(dbName, 0600, nil, backupBucketName, 1*time.Minute)
	if err != nil {
		log.Fatalln("Couldn't open db", err)
	}
	defer db.Close()

	// create main bucket
	bucketName := []byte("users")
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Fatalln("error creating bucket", err)
	}

	r := gin.Default()
	r.POST("/users", func(c *gin.Context) {
		u := &User{}
		if c.BindJSON(u) == nil {
			if u.Name == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
				return
			}
			err := db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket(bucketName)
				id, _ := b.NextSequence()
				u.ID = int(id)

				// Marshal user data into bytes.
				buf, err := json.Marshal(u)
				if err != nil {
					return err
				}

				// Persist bytes to users bucket.
				return b.Put(itob(u.ID), buf)
			})
			if err != nil {
				log.Println("Error saving user!", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Uh oh, error."})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "user saved"})
		}
	})
	r.GET("/users", func(c *gin.Context) {
		users := []*User{}
		err := db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := tx.Bucket(bucketName)

			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				fmt.Printf("key=%s, value=%s\n", k, v)
				u := &User{}
				err := json.Unmarshal(v, u)
				if err != nil {
					return err
				}
				users = append(users, u)
			}

			return nil
		})
		if err != nil {
			log.Println("Error getting users!", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Uh oh, error."})
			return
		}
		c.JSON(200, users)
	})
	r.Run() // listen and server on 0.0.0.0:8080
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
