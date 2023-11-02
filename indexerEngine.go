package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"log"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"context"
	"sync"
	"github.com/otiai10/gosseract/v2"
	_ "net/http/pprof"
)


type File struct {
	FileName string
	FilePath string
	FileData string
}



const MaxWorkers = 12
const batchSize = 5

func worker(filesChan <-chan string, wg *sync.WaitGroup, collection *mongo.Collection, ctx context.Context, errChan chan<- error) {
    defer wg.Done()
    batch := make([]interface{}, 0, batchSize)
	// Initialize the gosseract client once per worker
	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("eng", "hin", "urd") // Set the languages once
    for path := range filesChan {
		fmt.Println("Processing file:", path)
		fileData, err := extractTextWithClient(client, path) // Use the persistent client
		if err != nil {
			errChan <- fmt.Errorf("error extracting text from %q: %w", path, err)
			continue
		}

		file := File{
			FileName: filepath.Base(path),
			FilePath: path,
			FileData: fileData,
		}

		batch = append(batch, file)

		if len(batch) >= batchSize {
			if _, err := collection.InsertMany(ctx, batch); err != nil {
				errChan <- fmt.Errorf("error inserting batch into MongoDB: %w", err)
			}
			batch = batch[:0]
		}
	}

    // Insert any remaining files
    if len(batch) > 0 {
        if _, err := collection.InsertMany(ctx, batch); err != nil {
            fmt.Printf("error inserting final batch into MongoDB: %v\n", err)
        }
    }
}

func indexerEngine(rootPath string) {
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        log.Fatal("MONGO_URI not set in environment")
    }

    // Set client options with write concern for better performance
    clientOptions := options.Client().ApplyURI(mongoURI).SetWriteConcern(writeconcern.New(writeconcern.W(1)))

    // Connect to MongoDB
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        if err = client.Disconnect(context.TODO()); err != nil {
            log.Fatal(err)
        }
    }()

    fmt.Println("Connected to MongoDB!")

    collection := client.Database("tesseract").Collection("images")

    filesChan := make(chan string, MaxWorkers)
    var wg sync.WaitGroup
    ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, MaxWorkers)

	go func() {
        for err := range errChan {
            log.Println("Error received:", err)
        }
    }()
    // Start workers
    for i := 0; i < MaxWorkers; i++ {
        wg.Add(1)
        go worker(filesChan, &wg, collection, ctx,errChan)
    }



    // Walking the directory structure and sending files to the channel
    walkErr := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            fmt.Printf("error accessing path %q: %v\n", path, err)
            return err
        }

        if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".jpeg") {
            filesChan <- path
        }
        return nil
    })

    if walkErr != nil {
        fmt.Printf("error walking the path %v: %v\n", rootPath, walkErr)
    }

    close(filesChan)
    wg.Wait()
	close(errChan) 
    cancel() // Cancel the context to free resources
}