package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/gitferry/zeitgeber/identity"
	"github.com/gitferry/zeitgeber/log"
	"github.com/gitferry/zeitgeber/transport"
)

var configFile = flag.String("config", "config.json", "Configuration file for zeitgeber replica. Defaults to config.json.")

// Config contains every system configuration
type Config struct {
	Addrs     map[identity.NodeID]string `json:"address"`      // address for node communication
	HTTPAddrs map[identity.NodeID]string `json:"http_address"` // address for client server communication

	Policy    string  `json:"policy"`    // leader change policy {consecutive, majority}
	Threshold float64 `json:"threshold"` // threshold for policy in WPaxos {n consecutive or time interval in ms}

	Thrifty        bool    `json:"thrifty"`          // only send messages to a quorum
	BufferSize     int     `json:"buffer_size"`      // buffer size for maps
	ChanBufferSize int     `json:"chan_buffer_size"` // buffer size for channels
	MultiVersion   bool    `json:"multiversion"`     // create multi-version database
	Timeout        int     `json:"timeout"`
	ByzNo          int     `json:"byzNo"`
	BSize          int     `json:"bsize"`
	Benchmark      Bconfig `json:"benchmark"` // benchmark configuration
	Delta          int     `json:"delta"`     // timeout, seconds

	// for future implementation
	// Batching bool `json:"batching"`
	// Consistency string `json:"consistency"`
	// Codec string `json:"codec"` // codec for message serialization between nodes

	n int // total number of nodes
	//z   int         // total number of zones
	//npz map[int]int // nodes per zone
}

// Bconfig holds all benchmark configuration
type Bconfig struct {
	T            int    // total number of running time in seconds
	N            int    // total number of requests
	K            int    // key sapce
	Throttle     int    // requests per second throttle, unused if 0
	Concurrency  int    // number of simulated clients
	Distribution string // distribution
	// rounds       int    // repeat in many rounds sequentially

	// conflict distribution
	Conflicts int // percentage of conflicting keys
	Min       int // min key

	// normal distribution
	Mu    float64 // mu of normal distribution
	Sigma float64 // sigma of normal distribution
	Move  bool    // moving average (mu) of normal distribution
	Speed int     // moving speed in milliseconds intervals per key
}

// Config is global configuration singleton generated by init() func below
var Configuration Config

func init() {
	Configuration = MakeDefaultConfig()
}

// GetConfig returns paxi package configuration
func GetConfig() Config {
	return Configuration
}

func GetTimer() time.Duration {
	return time.Duration(time.Duration(Configuration.Timeout) * time.Millisecond)
}

// Simulation enable go channel transportation to simulate distributed environment
func Simulation() {
	*transport.Scheme = "chan"
}

// MakeDefaultConfig returns Config object with few default values
// only used by init() and master
func MakeDefaultConfig() Config {
	return Config{
		Policy:         "consecutive",
		Threshold:      3,
		BufferSize:     1024,
		ChanBufferSize: 1024,
		MultiVersion:   false,
		//Benchmark:      DefaultBConfig(),
	}
}

// IDs returns all node ids
func (c Config) IDs() []identity.NodeID {
	ids := make([]identity.NodeID, 0)
	for id := range c.Addrs {
		ids = append(ids, id)
	}
	return ids
}

// N returns total number of nodes
func (c Config) N() int {
	return c.n
}

// Z returns total number of zones
//func (c Config) Z() int {
//	return c.z
//}

// String is implemented to print the Configuration
func (c Config) String() string {
	config, err := json.Marshal(c)
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(config)
}

// Load loads configuration from Configuration file in JSON format
func (c *Config) Load() {
	file, err := os.Open(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		log.Fatal(err)
	}

	// test
	for i := 1; i <= 16; i++ {
		id := identity.NewNodeID(i)
		port := strconv.Itoa(3000 + i)
		addr := "tcp://127.0.0.1:" + port
		portHttp := strconv.Itoa(9000 + i)
		addrHttp := "http://127.0.0.1:" + portHttp
		c.Addrs[id] = addr
		c.HTTPAddrs[id] = addrHttp
	}

	c.ByzNo = 1
	c.Threshold = 11

	c.n = len(c.Addrs)
	//c.npz = make(map[int]int)
	//for id := range c.Addrs {
	//	c.n++
	//	//c.npz[id.Zone()]++
	//}
	//c.z = len(c.npz)
}

// Save saves configuration to file in JSON format
func (c Config) Save() error {
	file, err := os.Create(*configFile)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	return encoder.Encode(c)
}

func (c Config) IsByzantine(id identity.NodeID) bool {
	return c.ByzNo >= id.Node()
}
