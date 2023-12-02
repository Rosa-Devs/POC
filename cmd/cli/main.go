package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Rosa-Devs/POC/src/manifest"
	"github.com/Rosa-Devs/POC/src/p2p"
	db "github.com/Rosa-Devs/POC/src/store"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "pubsub-chat-example"

func main() {
	database := flag.String("d", "", "use it to create databse manifest file")
	ManifestFile := flag.String("m", "", "set Manifets file")
	FolderName := flag.String("f", "", "set db folder name")
	flag.Parse()

	if *database != "" {
		manifest.GenereateManifest(*database, true)
		return
	}

	if *ManifestFile == "" {
		log.Println("Specifi a manifest file... -m")
		return
	}

	ctx := context.Background()

	// create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New()
	if err != nil {
		panic(err)
	}

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	// setup local mDNS discovery
	if err := setupDiscovery(h); err != nil {
		panic(err)
	}

	// use the nickname from the cli flag, or a default if blank

	manifetstData := manifest.ReadManifestFromFile(*ManifestFile)

	// !! GLOBAl DB MANAGER !!
	//CREATE DATABSE INSTANCE
	Drvier := db.DB{}
	//START DATABSE INSTANCE
	if *FolderName != "" {
		Drvier.Start(*FolderName)
	} else {
		Drvier.Start("test_db_1")
	}

	//CREATE TEST DB
	Drvier.CreateDb(manifetstData.Name)

	// !! WORKING WITH SPECIFIED BATABASE !!
	db1 := Drvier.GetDb(manifetstData.Name, ps, manifetstData, h.ID())
	db1.StartWorker()

	err = db1.CreatePool("test_pool")
	if err != nil {
		log.Println("Mayby this pool alredy exist:", err)
		//return
	}

	_, err = db1.GetPool("test_pool", true)
	if err != nil {
		log.Println(err)
		return
	}

	// go func() {
	// 	//SIMULATE ADDING DATA
	// 	rand.Seed(time.Now().UnixNano())
	// 	for {
	// 		// Generate random data
	// 		randomData := map[string]interface{}{
	// 			"field1": rand.Intn(100),             // Random integer between 0 and 100
	// 			"field2": rand.Float64() * 100,       // Random float between 0 and 100
	// 			"field3": uuid.New().String(),        // Random UUID as a string
	// 			"field4": time.Now().UnixNano(),      // Current timestamp in nanoseconds
	// 			"field5": fmt.Sprintf("Record%d", 1), // Custom string with record number
	// 		}

	// 		// Convert data to JSON
	// 		jsonData, err := json.Marshal(randomData)
	// 		if err != nil {
	// 			fmt.Println("Error marshaling JSON:", err)
	// 			return
	// 		}

	// 		// Call Record function to save the record
	// 		err = pool.Record(jsonData)
	// 		if err != nil {
	// 			fmt.Println("Error recording data:", err)
	// 			return
	// 		}
	// 		time.Sleep(time.Millisecond * 10)
	// 	}
	// }()

	// go func() {
	// 	for {
	// 		filter := map[string]interface{}{
	// 			"field1": 96, // Random integer between 0 and 100
	// 		}

	// 		data, err := pool.Filter(filter)
	// 		if err != nil {
	// 			fmt.Println("Data:", data)
	// 			fmt.Println("Error filtering data:", err)
	// 		}
	// 		log.Println(data)
	// 		time.Sleep(time.Millisecond * 70)
	// 	}
	// }()

	// go func() {

	// 	var prevHashTree map[string]string
	// 	for {
	// 		startTime := time.Now()

	// 		// Calculate the current hash tree
	// 		currentRoot, currentHashTree, err := pool.GenereateMerkleTree()
	// 		if err != nil {
	// 			println("Error calculating current hash tree:", err)
	// 			continue
	// 		}

	// 		changedFiles := pool.CalculateChangedFiles(prevHashTree, currentHashTree)

	// 		// Show or log the changed files
	// 		showChangedFiles(changedFiles)

	// 		// Update the previous hash tree for the next iteration
	// 		prevHashTree = currentHashTree

	// 		endTime := time.Now()
	// 		duration := endTime.Sub(startTime)
	// 		log.Printf("Root: %s, Time: %s", currentRoot, duration)

	// 		time.Sleep(time.Second) // Adjust the sleep duration as needed
	// 	}
	// }()

	// go func() {
	// 	for {
	// 		db1.PublishUpdate(db.Action{
	// 			Channel:  manifetstData.PubSub,
	// 			SenderID: "root",
	// 			Data: db.Data{
	// 				FileID:  "1",
	// 				Content: []byte("root"),
	// 			},
	// 			Type: db.Update,
	// 		})
	// 	}
	// }()

	for {
	}

}

// func showChangedFiles(changedFiles []string) {
// 	if len(changedFiles) > 0 {
// 		// Print or log the changed files
// 		for _, filePath := range changedFiles {
// 			println("Changed file:", filePath)
// 		}
// 	} else {
// 		///println("No files have changed.")
// 	}
// }

// printErr is like fmt.Printf, but writes to stderr.
func printErr(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}

// defaultNick generates a nickname based on the $USER environment variable and
// the last 8 chars of a peer ID.
func defaultNick(p peer.ID) string {
	return fmt.Sprintf("%s-%s", os.Getenv("USER"), p2p.ShortID(p))
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID)
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID, err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}
