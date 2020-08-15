## kvgo

An embeddable, persistent and distributed reliable Key-Value database engine.

## Features

* support both embedded and server-client running modes.
* fast and persistent key-value storage engine based on [goleveldb](https://github.com/syndtr/goleveldb).
* data is stored sorted by key, forward and backward query is supported over the data.
* data is automatically compressed using the snappy.
* support paxos-based distributed deployment and provide service via gRPC.
* friendly and easy way to run cluster mode in daemon and systemd ([kvgo-server](https://github.com/lynkdb/kvgo-server)).
* mount a kvgo database as a FUSE filesystem [kvgo-fs-mount](https://github.com/lynkdb/kvgo-fs-mount).


## Getting Started

### Installing

``` shell
go get -u github.com/lynkdb/kvgo
```

### Opening a database in embedded mode

``` go
package main

import (
	"fmt"

	"github.com/lynkdb/kvgo"
)

func main() {

	db, err := kvgo.Open(kvgo.ConfigStorage{
		DataDirectory: "/tmp/kvgo-demo",
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if rs := db.NewWriter([]byte("key"), []byte("value")).Commit(); rs.OK() {
		fmt.Println("OK")
	} else {
		fmt.Println("ER", rs.Message)
	}
}
```

### Opening a database in server-client mode

``` go
package main

import (
	"fmt"

	"github.com/hooto/hflag4g/hflag"
	"github.com/lynkdb/kvgo"
)

var (
	addr      = "127.0.0.1:9100"
	accessKey = kvgo.NewSystemAccessKey()
	Server    *kvgo.Conn
	tlsCert   *kvgo.ConfigTLSCertificate
	err       error
)

func main() {

	if _, ok := hflag.ValueOK("tls_enable"); ok {
		// openssl genrsa -out server.key 2048
		// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650 -subj '/CN=CommonName'
		tlsCert = &kvgo.ConfigTLSCertificate{
			ServerKeyFile:  "server.key",
			ServerCertFile: "server.crt",
		}
	}

	if err := startServer(); err != nil {
		panic(err)
	}

	client()
}

func startServer() error {

	if Server, err = kvgo.Open(kvgo.ConfigStorage{
		DataDirectory: "/tmp/kvgo-server",
	}, kvgo.ConfigServer{
		Bind:        addr,
		AccessKey:   accessKey,
		AuthTLSCert: tlsCert,
	}); err != nil {
		return err
	}

	return nil
}

func client() {

	clientConfig := kvgo.ClientConfig{
		Addr:        addr,
		AccessKey:   accessKey,
		AuthTLSCert: tlsCert,
	}

	client, err := clientConfig.NewClient()
	if err != nil {
		panic(err)
	}

	if rs := client.NewWriter([]byte("demo-key"), []byte("demo-value")).Commit(); rs.OK() {
		fmt.Println("OK")
	} else {
		fmt.Println("ER", rs.Message)
	}
}
```

### Deployment in distributed reliable database cluster mode

``` go
package main

import (
	"fmt"

	"github.com/lynkdb/kvgo"
)

var (
	accessKey = kvgo.NewSystemAccessKey()
	mainNodes = []*kvgo.ClientConfig{
		{
			Addr:      "127.0.0.1:9101",
			AccessKey: accessKey,
		},
		{
			Addr:      "127.0.0.1:9102",
			AccessKey: accessKey,
		},
		{
			Addr:      "127.0.0.1:9103",
			AccessKey: accessKey,
		},
	}
)

func main() {

	if err := startCluster(); err != nil {
		panic(err)
	}

	client()
}

func startCluster() error {

	for i, m := range mainNodes {

		_, err := kvgo.Open(kvgo.ConfigStorage{
			DataDirectory: fmt.Sprintf("/tmp/kvgo-cluster-%d", i),
		}, kvgo.ConfigServer{
			Bind:      m.Addr,
			AccessKey: accessKey,
		}, kvgo.ConfigCluster{
			MainNodes: mainNodes,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func client() {

	clientConfig := kvgo.ClientConfig{
		Addr:      "127.0.0.1:9102",
		AccessKey: accessKey,
	}

	client, err := clientConfig.NewClient()
	if err != nil {
		panic(err)
	}

	if rs := client.NewWriter([]byte("demo-key"), []byte("demo-value")).Commit(); rs.OK() {
		fmt.Println("OK")
	} else {
		fmt.Println("ER", rs.Message)
	}
}
```

Tips: use [kvgo-server](https://github.com/lynkdb/kvgo-server) to deploy the cluster in daemon and systemd.


## Data Write/Read APIs

### Write Key-Value Data into database

``` go
var (
	key = []byte("demo-key")
	val = []byte("demo-value-data")
)

# general method
if rs := db.NewWriter(key, val).Commit(); rs.OK() {
	fmt.Println("OK")
}

# write the value of a key, only if the key does not exist
if rs := db.NewWriter(key, val).ModeCreateSet(true).Commit(); rs.OK() {
	fmt.Println("OK")
}

# set a timeout on key. After the timeout has expired, the key will automatically be deleted.
# timeout setup in milliseconds
if rs := db.NewWriter(key, val).ExpireSet(3000).Commit(); rs.OK() {
	fmt.Println("OK")
}

# delete key
if rs := db.NewWriter(key).ModeDeleteSet(true).Commit(); rs.OK() {
	fmt.Println("OK")
}

# write the new-value of a key, only if the key exist and the old value is the same as the commited check value.
if rs := db.NewWriter(key, "new-value").PrevDataCheckSet("old-value").Commit(); rs.OK() {
	fmt.Println("OK")
}

# write the value of a key and automatically create a auto-increment meta value.
# the auto-increment value will keep the same if the key exist.
if rs := db.NewWriter(key, "value").IncrNamespaceSet("def").Commit(); rs.OK() {
	fmt.Println("OK". rs.Meta.IncrId)
}

# multi value type supports
rs = db.NewWriter(key, []byte("value")).Commit()
rs = db.NewWriter(key, "value").Commit()
rs = db.NewWriter(key, 1.01).Commit()
type StructObject struct {
	Name string
}
obj := StructObject{
	Name: "test",
}
rs = db.NewWriter(key, obj).Commit()

```

### Read or Query Key-Value Data from database

``` go
var (
	key = []byte("demo-key")
)

# query one key-value item
if rs := db.NewReader(key).Query(); rs.OK() {
	fmt.Println("OK", rs.DataValue().String())
} else if rs.NotFound() {
	fmt.Println("Not Found")
} else {
	fmt.Println("Error", rs.Message)
}

# query multi key-value items from a key-range in forward way
if rs := db.NewReader().
	KeyRangeSet([]byte("00"), []byte("zz")).
	LimitNumSet(10).Query(); rs.OK() {
	for i, item := range rs.Items {
		fmt.Printf("N %d, Value %s\n", i, item.DataValue().String())
	}
}

# query multi key-value items from a key-range in backward way
if rs := db.NewReader().
	KeyRangeSet([]byte("zz"), []byte("00")).
	ModeRevRangeSet(true).
	LimitNumSet(10).Query(); rs.OK() {
	for i, item := range rs.Items {
		fmt.Printf("N %d, Value %s\n", i, item.DataValue().String())
	}
}

# query multi key-value items by the paxos-based auto-increment log version.
lastVersion := uint64(0)
if rs := db.NewReader().
	LogOffsetSet(lastVersion).
	LimitNumSet(10).Query(); rs.OK() {
	for i, item := range rs.Items {
		lastVersion = item.Meta.Version
		fmt.Printf("N %d, Version \n", i, lastVersion)
	}
}

# get value data in multi types
_ = rs.DataValue().Bytes()
_ = rs.DataValue().String()
_ = rs.DataValue().Int() # or Int8(), Int16(), Int32(), Int64()
_ = rs.DataValue().Uint() # or Uint8(), Uint16(), Uint32(), Int64()
_ = rs.DataValue().Bool()
_ = rs.DataValue().Float64()
type StructObject struct {
	Name string
}
var item StructObject
if err := rs.DataValue().Decode(&item, nil); err == nil {
	// ...
}
```

## Performance


### test environment

* CPU: Intel i7-7700 CPU @ 3.60GHz (4 cores, 8 threads)
* SSD: Intel 760P 512GB M.2/NVMe
* OS: CentOS 7.7.1908 x86_64
* kvgo: version 0.2.0 (write_buffer 64MB, block_cache_size 64MB)
* redis: version 5.0.7 (disable save the DB on disk)
* data keys: 40 bytes each
* data values: 1024 bytes each

### typical performance in embed, 1 node and 3 nodes modes:

![typical-benchmark](bench/kvgo_throughput_avg.svg)


### kvgo vs redis in 1 node mode:

![kvgo-vs-redis-benchmark](bench/kvgo_redis_throughput_avg.svg)



## Dependent or referenced

* leveldb [https://github.com/google/leveldb](https://github.com/google/leveldb)
* goleveldb [github.com/syndtr/goleveldb](github.com/syndtr/goleveldb)
* snappy [http://github.com/google/snappy](http://github.com/google/snappy)
* snappy in go [https://github.com/golang/snappy](https://github.com/golang/snappy)
