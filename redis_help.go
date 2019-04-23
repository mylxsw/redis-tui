package main

import (
	"strings"
)

type RedisHelp struct {
	Command string
	Args    string
	Version string
	Desc    string
}

func RedisMatchedCommands(text string) []RedisHelp {
	value := strings.ToUpper(text)
	segments := strings.SplitN(value, " ", 3)

	var matched = make([]RedisHelp, 0)
	for _, help := range commandHelps {
		cmd := strings.SplitN(help.Command, " ", 2)
		// if segments[0] != cmd[0] {
		// 	continue
		// }
		//
		// if len(cmd) > 1 && len(segments) > 1 && !strings.HasPrefix(cmd[1], segments[1]) {
		// 	continue
		// }

		if len(cmd) > 1 && len(segments) > 1 {
			if help.Command == segments[0]+" "+segments[1] {
				matched = []RedisHelp{help}
				break
			}
			if strings.HasPrefix(help.Command, segments[0]+" "+segments[1]) {
				matched = append(matched, help)
				continue
			}
		} else {
			if help.Command == segments[0] {
				matched = []RedisHelp{help}
				break
			}

			if strings.HasPrefix(help.Command, segments[0]) {
				matched = append(matched, help)
				continue
			}
		}

	}

	return matched
}

func RedisHelpMatch(text string, matchedFunc func(help RedisHelp)) bool {
	value := strings.ToUpper(text)
	segments := strings.SplitN(value, " ", 3)

	var matchedIndex = -1
	for i, help := range commandHelps {
		cmd := strings.SplitN(help.Command, " ", 2)
		if segments[0] != cmd[0] {
			continue
		}

		if len(cmd) > 1 && len(segments) > 1 && !strings.HasPrefix(cmd[1], segments[1]) {
			continue
		}

		matchedIndex = i
		break
	}

	if matchedIndex > -1 {
		help := commandHelps[matchedIndex]
		matchedFunc(help)
	}

	return matchedIndex > -1
}

var commandHelps = []RedisHelp{
	{Command: "APPEND", Args: "key value", Version: "2.0.0", Desc: "Append a value to a key"},
	{Command: "AUTH", Args: "password", Version: "1.0.0", Desc: "Authenticate to the server"},
	{Command: "BGREWRITEAOF", Args: "-", Version: "1.0.0", Desc: "Asynchronously rewrite the append-only file"},
	{Command: "BGSAVE", Args: "-", Version: "1.0.0", Desc: "Asynchronously save the dataset to disk"},
	{Command: "BITCOUNT", Args: "key [start end]", Version: "2.6.0", Desc: "Count set bits in a string"},
	{Command: "BITFIELD", Args: "key [GET type offset] [SET type offset value] [INCRBY type offset increment] [OVERFLOW WRAP|SAT|FAIL]", Version: "3.2.0", Desc: "Perform arbitrary bitfield integer operations on strings"},
	{Command: "BITOP", Args: "operation destkey key [key ...]", Version: "2.6.0", Desc: "Perform bitwise operations between strings"},
	{Command: "BITPOS", Args: "key bit [start] [end]", Version: "2.8.7", Desc: "Find first bit set or clear in a string"},
	{Command: "BLPOP", Args: "key [key ...] timeout", Version: "2.0.0", Desc: "Remove and get the first element in a list, or block until one is available"},
	{Command: "BRPOP", Args: "key [key ...] timeout", Version: "2.0.0", Desc: "Remove and get the last element in a list, or block until one is available"},
	{Command: "BRPOPLPUSH", Args: "source destination timeout", Version: "2.2.0", Desc: "Pop a value from a list, push it to another list and return it; or block until one is available"},
	{Command: "BZPOPMAX", Args: "key [key ...] timeout", Version: "5.0.0", Desc: "Remove and return the member with the highest score from one or more sorted sets, or block until one is available"},
	{Command: "BZPOPMIN", Args: "key [key ...] timeout", Version: "5.0.0", Desc: "Remove and return the member with the lowest score from one or more sorted sets, or block until one is available"},
	{Command: "CLIENT GETNAME", Args: "-", Version: "2.6.9", Desc: "Get the current connection name"},
	{Command: "CLIENT ID", Args: "-", Version: "5.0.0", Desc: "Returns the client ID for the current connection"},
	{Command: "CLIENT KILL", Args: "[ip:port] [ID client-id] [TYPE normal|master|slave|pubsub] [ADDR ip:port] [SKIPME yes/no]", Version: "2.4.0", Desc: "Kill the connection of a client"},
	{Command: "CLIENT LIST", Args: "-", Version: "2.4.0", Desc: "Get the list of client connections"},
	{Command: "CLIENT PAUSE", Args: "timeout", Version: "2.9.50", Desc: "Stop processing commands from clients for some time"},
	{Command: "CLIENT REPLY", Args: "ON|OFF|SKIP", Version: "3.2", Desc: "Instruct the server whether to reply to commands"},
	{Command: "CLIENT SETNAME", Args: "connection-name", Version: "2.6.9", Desc: "Set the current connection name"},
	{Command: "CLIENT UNBLOCK", Args: "client-id [TIMEOUT|ERROR]", Version: "5.0.0", Desc: "Unblock a client blocked in a blocking command from a different connection"},
	{Command: "CLUSTER ADDSLOTS", Args: "slot [slot ...]", Version: "3.0.0", Desc: "Assign new hash slots to receiving node"},
	{Command: "CLUSTER COUNT-FAILURE-REPORTS", Args: "node-id", Version: "3.0.0", Desc: "Return the number of failure reports active for a given node"},
	{Command: "CLUSTER COUNTKEYSINSLOT", Args: "slot", Version: "3.0.0", Desc: "Return the number of local keys in the specified hash slot"},
	{Command: "CLUSTER DELSLOTS", Args: "slot [slot ...]", Version: "3.0.0", Desc: "Set hash slots as unbound in receiving node"},
	{Command: "CLUSTER FAILOVER", Args: "[FORCE|TAKEOVER]", Version: "3.0.0", Desc: "Forces a replica to perform a manual failover of its master."},
	{Command: "CLUSTER FORGET", Args: "node-id", Version: "3.0.0", Desc: "Remove a node from the nodes table"},
	{Command: "CLUSTER GETKEYSINSLOT", Args: "slot count", Version: "3.0.0", Desc: "Return local key names in the specified hash slot"},
	{Command: "CLUSTER INFO", Args: "-", Version: "3.0.0", Desc: "Provides info about Redis Cluster node state"},
	{Command: "CLUSTER KEYSLOT", Args: "key", Version: "3.0.0", Desc: "Returns the hash slot of the specified key"},
	{Command: "CLUSTER MEET", Args: "ip port", Version: "3.0.0", Desc: "Force a node cluster to handshake with another node"},
	{Command: "CLUSTER NODES", Args: "-", Version: "3.0.0", Desc: "Get Cluster config for the node"},
	{Command: "CLUSTER REPLICAS", Args: "node-id", Version: "5.0.0", Desc: "List replica nodes of the specified master node"},
	{Command: "CLUSTER REPLICATE", Args: "node-id", Version: "3.0.0", Desc: "Reconfigure a node as a replica of the specified master node"},
	{Command: "CLUSTER RESET", Args: "[HARD|SOFT]", Version: "3.0.0", Desc: "Reset a Redis Cluster node"},
	{Command: "CLUSTER SAVECONFIG", Args: "-", Version: "3.0.0", Desc: "Forces the node to save cluster state on disk"},
	{Command: "CLUSTER SET-CONFIG-EPOCH", Args: "config-epoch", Version: "3.0.0", Desc: "Set the configuration epoch in a new node"},
	{Command: "CLUSTER SETSLOT", Args: "slot IMPORTING|MIGRATING|STABLE|NODE [node-id]", Version: "3.0.0", Desc: "Bind a hash slot to a specific node"},
	{Command: "CLUSTER SLAVES", Args: "node-id", Version: "3.0.0", Desc: "List replica nodes of the specified master node"},
	{Command: "CLUSTER SLOTS", Args: "-", Version: "3.0.0", Desc: "Get array of Cluster slot to node mappings"},
	{Command: "COMMAND", Args: "-", Version: "2.8.13", Desc: "Get array of Redis command details"},
	{Command: "COMMAND COUNT", Args: "-", Version: "2.8.13", Desc: "Get total number of Redis commands"},
	{Command: "COMMAND GETKEYS", Args: "-", Version: "2.8.13", Desc: "Extract keys given a full Redis command"},
	{Command: "COMMAND INFO", Args: "command-name [command-name ...]", Version: "2.8.13", Desc: "Get array of specific Redis command details"},
	{Command: "CONFIG GET", Args: "parameter", Version: "2.0.0", Desc: "Get the value of a configuration parameter"},
	{Command: "CONFIG RESETSTAT", Args: "-", Version: "2.0.0", Desc: "Reset the stats returned by INFO"},
	{Command: "CONFIG REWRITE", Args: "-", Version: "2.8.0", Desc: "Rewrite the configuration file with the in memory configuration"},
	{Command: "CONFIG SET", Args: "parameter value", Version: "2.0.0", Desc: "Set a configuration parameter to the given value"},
	{Command: "DBSIZE", Args: "-", Version: "1.0.0", Desc: "Return the number of keys in the selected database"},
	{Command: "DEBUG OBJECT", Args: "key", Version: "1.0.0", Desc: "Get debugging information about a key"},
	{Command: "DEBUG SEGFAULT", Args: "-", Version: "1.0.0", Desc: "Make the server crash"},
	{Command: "DECR", Args: "key", Version: "1.0.0", Desc: "Decrement the integer value of a key by one"},
	{Command: "DECRBY", Args: "key decrement", Version: "1.0.0", Desc: "Decrement the integer value of a key by the given number"},
	{Command: "DEL", Args: "key [key ...]", Version: "1.0.0", Desc: "Delete a key"},
	{Command: "DISCARD", Args: "-", Version: "2.0.0", Desc: "Discard all commands issued after MULTI"},
	{Command: "DUMP", Args: "key", Version: "2.6.0", Desc: "Return a serialized version of the value stored at the specified key."},
	{Command: "ECHO", Args: "message", Version: "1.0.0", Desc: "Echo the given string"},
	{Command: "EVAL", Args: "script numkeys key [key ...] arg [arg ...]", Version: "2.6.0", Desc: "Execute a Lua script server side"},
	{Command: "EVALSHA", Args: "sha1 numkeys key [key ...] arg [arg ...]", Version: "2.6.0", Desc: "Execute a Lua script server side"},
	{Command: "EXEC", Args: "-", Version: "1.2.0", Desc: "Execute all commands issued after MULTI"},
	{Command: "EXISTS", Args: "key [key ...]", Version: "1.0.0", Desc: "Determine if a key exists"},
	{Command: "EXPIRE", Args: "key seconds", Version: "1.0.0", Desc: "Set a key's time to live in seconds"},
	{Command: "EXPIREAT", Args: "key timestamp", Version: "1.2.0", Desc: "Set the expiration for a key as a UNIX timestamp"},
	{Command: "FLUSHALL", Args: "[ASYNC]", Version: "1.0.0", Desc: "Remove all keys from all databases"},
	{Command: "FLUSHDB", Args: "[ASYNC]", Version: "1.0.0", Desc: "Remove all keys from the current database"},
	{Command: "GEOADD", Args: "key longitude latitude member [longitude latitude member ...]", Version: "3.2.0", Desc: "Add one or more geospatial items in the geospatial index represented using a sorted set"},
	{Command: "GEODIST", Args: "key member1 member2 [unit]", Version: "3.2.0", Desc: "Returns the distance between two members of a geospatial index"},
	{Command: "GEOHASH", Args: "key member [member ...]", Version: "3.2.0", Desc: "Returns members of a geospatial index as standard geohash strings"},
	{Command: "GEOPOS", Args: "key member [member ...]", Version: "3.2.0", Desc: "Returns longitude and latitude of members of a geospatial index"},
	{Command: "GEORADIUS", Args: "key longitude latitude radius m|km|ft|mi [WITHCOORD] [WITHDIST] [WITHHASH] [COUNT count] [ASC|DESC] [STORE key] [STOREDIST key]", Version: "3.2.0", Desc: "Query a sorted set representing a geospatial index to fetch members matching a given maximum distance from a point"},
	{Command: "GEORADIUSBYMEMBER", Args: "key member radius m|km|ft|mi [WITHCOORD] [WITHDIST] [WITHHASH] [COUNT count] [ASC|DESC] [STORE key] [STOREDIST key]", Version: "3.2.0", Desc: "Query a sorted set representing a geospatial index to fetch members matching a given maximum distance from a member"},
	{Command: "GET", Args: "key", Version: "1.0.0", Desc: "Get the value of a key"},
	{Command: "GETBIT", Args: "key offset", Version: "2.2.0", Desc: "Returns the bit value at offset in the string value stored at key"},
	{Command: "GETRANGE", Args: "key start end", Version: "2.4.0", Desc: "Get a substring of the string stored at a key"},
	{Command: "GETSET", Args: "key value", Version: "1.0.0", Desc: "Set the string value of a key and return its old value"},
	{Command: "HDEL", Args: "key field [field ...]", Version: "2.0.0", Desc: "Delete one or more hash fields"},
	{Command: "HEXISTS", Args: "key field", Version: "2.0.0", Desc: "Determine if a hash field exists"},
	{Command: "HGET", Args: "key field", Version: "2.0.0", Desc: "Get the value of a hash field"},
	{Command: "HGETALL", Args: "key", Version: "2.0.0", Desc: "Get all the fields and values in a hash"},
	{Command: "HINCRBY", Args: "key field increment", Version: "2.0.0", Desc: "Increment the integer value of a hash field by the given number"},
	{Command: "HINCRBYFLOAT", Args: "key field increment", Version: "2.6.0", Desc: "Increment the float value of a hash field by the given amount"},
	{Command: "HKEYS", Args: "key", Version: "2.0.0", Desc: "Get all the fields in a hash"},
	{Command: "HLEN", Args: "key", Version: "2.0.0", Desc: "Get the number of fields in a hash"},
	{Command: "HMGET", Args: "key field [field ...]", Version: "2.0.0", Desc: "Get the values of all the given hash fields"},
	{Command: "HMSET", Args: "key field value [field value ...]", Version: "2.0.0", Desc: "Set multiple hash fields to multiple values"},
	{Command: "HSCAN", Args: "key cursor [MATCH pattern] [COUNT count]", Version: "2.8.0", Desc: "Incrementally iterate hash fields and associated values"},
	{Command: "HSET", Args: "key field value", Version: "2.0.0", Desc: "Set the string value of a hash field"},
	{Command: "HSETNX", Args: "key field value", Version: "2.0.0", Desc: "Set the value of a hash field, only if the field does not exist"},
	{Command: "HSTRLEN", Args: "key field", Version: "3.2.0", Desc: "Get the length of the value of a hash field"},
	{Command: "HVALS", Args: "key", Version: "2.0.0", Desc: "Get all the values in a hash"},
	{Command: "INCR", Args: "key", Version: "1.0.0", Desc: "Increment the integer value of a key by one"},
	{Command: "INCRBY", Args: "key increment", Version: "1.0.0", Desc: "Increment the integer value of a key by the given amount"},
	{Command: "INCRBYFLOAT", Args: "key increment", Version: "2.6.0", Desc: "Increment the float value of a key by the given amount"},
	{Command: "INFO", Args: "[section]", Version: "1.0.0", Desc: "Get information and statistics about the server"},
	{Command: "KEYS", Args: "pattern", Version: "1.0.0", Desc: "Find all keys matching the given pattern"},
	{Command: "LASTSAVE", Args: "-", Version: "1.0.0", Desc: "Get the UNIX time stamp of the last successful save to disk"},
	{Command: "LINDEX", Args: "key index", Version: "1.0.0", Desc: "Get an element from a list by its index"},
	{Command: "LINSERT", Args: "key BEFORE|AFTER pivot value", Version: "2.2.0", Desc: "Insert an element before or after another element in a list"},
	{Command: "LLEN", Args: "key", Version: "1.0.0", Desc: "Get the length of a list"},
	{Command: "LPOP", Args: "key", Version: "1.0.0", Desc: "Remove and get the first element in a list"},
	{Command: "LPUSH", Args: "key value [value ...]", Version: "1.0.0", Desc: "Prepend one or multiple values to a list"},
	{Command: "LPUSHX", Args: "key value", Version: "2.2.0", Desc: "Prepend a value to a list, only if the list exists"},
	{Command: "LRANGE", Args: "key start stop", Version: "1.0.0", Desc: "Get a range of elements from a list"},
	{Command: "LREM", Args: "key count value", Version: "1.0.0", Desc: "Remove elements from a list"},
	{Command: "LSET", Args: "key index value", Version: "1.0.0", Desc: "Set the value of an element in a list by its index"},
	{Command: "LTRIM", Args: "key start stop", Version: "1.0.0", Desc: "Trim a list to the specified range"},
	{Command: "MEMORY DOCTOR", Args: "-", Version: "4.0.0", Desc: "Outputs memory problems report"},
	{Command: "MEMORY HELP", Args: "-", Version: "4.0.0", Desc: "Show helpful text about the different subcommands"},
	{Command: "MEMORY MALLOC-STATS", Args: "-", Version: "4.0.0", Desc: "Show allocator internal stats"},
	{Command: "MEMORY PURGE", Args: "-", Version: "4.0.0", Desc: "Ask the allocator to release memory"},
	{Command: "MEMORY STATS", Args: "-", Version: "4.0.0", Desc: "Show memory usage details"},
	{Command: "MEMORY USAGE", Args: "key [SAMPLES count]", Version: "4.0.0", Desc: "Estimate the memory usage of a key"},
	{Command: "MGET", Args: "key [key ...]", Version: "1.0.0", Desc: "Get the values of all the given keys"},
	{Command: "MIGRATE", Args: "host port key | destination-db timeout [COPY] [REPLACE] [KEYS key]", Version: "2.6.0", Desc: "Atomically transfer a key from a Redis instance to another one."},
	{Command: "MONITOR", Args: "-", Version: "1.0.0", Desc: "Listen for all requests received by the server in real time"},
	{Command: "MOVE", Args: "key db", Version: "1.0.0", Desc: "Move a key to another database"},
	{Command: "MSET", Args: "key value [key value ...]", Version: "1.0.1", Desc: "Set multiple keys to multiple values"},
	{Command: "MSETNX", Args: "key value [key value ...]", Version: "1.0.1", Desc: "Set multiple keys to multiple values, only if none of the keys exist"},
	{Command: "MULTI", Args: "-", Version: "1.2.0", Desc: "Mark the start of a transaction block"},
	{Command: "OBJECT", Args: "subcommand [arguments [arguments ...]]", Version: "2.2.3", Desc: "Inspect the internals of Redis objects"},
	{Command: "PERSIST", Args: "key", Version: "2.2.0", Desc: "Remove the expiration from a key"},
	{Command: "PEXPIRE", Args: "key milliseconds", Version: "2.6.0", Desc: "Set a key's time to live in milliseconds"},
	{Command: "PEXPIREAT", Args: "key milliseconds-timestamp", Version: "2.6.0", Desc: "Set the expiration for a key as a UNIX timestamp specified in milliseconds"},
	{Command: "PFADD", Args: "key element [element ...]", Version: "2.8.9", Desc: "Adds the specified elements to the specified HyperLogLog."},
	{Command: "PFCOUNT", Args: "key [key ...]", Version: "2.8.9", Desc: "Return the approximated cardinality of the set(s) observed by the HyperLogLog at key(s)."},
	{Command: "PFMERGE", Args: "destkey sourcekey [sourcekey ...]", Version: "2.8.9", Desc: "Merge N different HyperLogLogs into a single one."},
	{Command: "PING", Args: "[message]", Version: "1.0.0", Desc: "Ping the server"},
	{Command: "PSETEX", Args: "key milliseconds value", Version: "2.6.0", Desc: "Set the value and expiration in milliseconds of a key"},
	{Command: "PSUBSCRIBE", Args: "pattern [pattern ...]", Version: "2.0.0", Desc: "Listen for messages published to channels matching the given patterns"},
	{Command: "PTTL", Args: "key", Version: "2.6.0", Desc: "Get the time to live for a key in milliseconds"},
	{Command: "PUBLISH", Args: "channel message", Version: "2.0.0", Desc: "Post a message to a channel"},
	{Command: "PUBSUB", Args: "subcommand [argument [argument ...]]", Version: "2.8.0", Desc: "Inspect the state of the Pub/Sub subsystem"},
	{Command: "PUNSUBSCRIBE", Args: "[pattern [pattern ...]]", Version: "2.0.0", Desc: "Stop listening for messages posted to channels matching the given patterns"},
	{Command: "QUIT", Args: "-", Version: "1.0.0", Desc: "Close the connection"},
	{Command: "RANDOMKEY", Args: "-", Version: "1.0.0", Desc: "Return a random key from the keyspace"},
	{Command: "READONLY", Args: "-", Version: "3.0.0", Desc: "Enables read queries for a connection to a cluster replica node"},
	{Command: "READWRITE", Args: "-", Version: "3.0.0", Desc: "Disables read queries for a connection to a cluster replica node"},
	{Command: "RENAME", Args: "key newkey", Version: "1.0.0", Desc: "Rename a key"},
	{Command: "RENAMENX", Args: "key newkey", Version: "1.0.0", Desc: "Rename a key, only if the new key does not exist"},
	{Command: "REPLICAOF", Args: "host port", Version: "5.0.0", Desc: "Make the server a replica of another instance, or promote it as master."},
	{Command: "RESTORE", Args: "key ttl serialized-value [REPLACE]", Version: "2.6.0", Desc: "Create a key using the provided serialized value, previously obtained using DUMP."},
	{Command: "ROLE", Args: "-", Version: "2.8.12", Desc: "Return the role of the instance in the context of replication"},
	{Command: "RPOP", Args: "key", Version: "1.0.0", Desc: "Remove and get the last element in a list"},
	{Command: "RPOPLPUSH", Args: "source destination", Version: "1.2.0", Desc: "Remove the last element in a list, prepend it to another list and return it"},
	{Command: "RPUSH", Args: "key value [value ...]", Version: "1.0.0", Desc: "Append one or multiple values to a list"},
	{Command: "RPUSHX", Args: "key value", Version: "2.2.0", Desc: "Append a value to a list, only if the list exists"},
	{Command: "SADD", Args: "key member [member ...]", Version: "1.0.0", Desc: "Add one or more members to a set"},
	{Command: "SAVE", Args: "-", Version: "1.0.0", Desc: "Synchronously save the dataset to disk"},
	{Command: "SCAN", Args: "cursor [MATCH pattern] [COUNT count]", Version: "2.8.0", Desc: "Incrementally iterate the keys space"},
	{Command: "SCARD", Args: "key", Version: "1.0.0", Desc: "Get the number of members in a set"},
	{Command: "SCRIPT DEBUG", Args: "YES|SYNC|NO", Version: "3.2.0", Desc: "Set the debug mode for executed scripts."},
	{Command: "SCRIPT EXISTS", Args: "sha1 [sha1 ...]", Version: "2.6.0", Desc: "Check existence of scripts in the script cache."},
	{Command: "SCRIPT FLUSH", Args: "-", Version: "2.6.0", Desc: "Remove all the scripts from the script cache."},
	{Command: "SCRIPT KILL", Args: "-", Version: "2.6.0", Desc: "Kill the script currently in execution."},
	{Command: "SCRIPT LOAD", Args: "script", Version: "2.6.0", Desc: "Load the specified Lua script into the script cache."},
	{Command: "SDIFF", Args: "key [key ...]", Version: "1.0.0", Desc: "Subtract multiple sets"},
	{Command: "SDIFFSTORE", Args: "destination key [key ...]", Version: "1.0.0", Desc: "Subtract multiple sets and store the resulting set in a key"},
	{Command: "SELECT", Args: "index", Version: "1.0.0", Desc: "Change the selected database for the current connection"},
	{Command: "SET", Args: "key value [expiration EX seconds|PX milliseconds] [NX|XX]", Version: "1.0.0", Desc: "Set the string value of a key"},
	{Command: "SETBIT", Args: "key offset value", Version: "2.2.0", Desc: "Sets or clears the bit at offset in the string value stored at key"},
	{Command: "SETEX", Args: "key seconds value", Version: "2.0.0", Desc: "Set the value and expiration of a key"},
	{Command: "SETNX", Args: "key value", Version: "1.0.0", Desc: "Set the value of a key, only if the key does not exist"},
	{Command: "SETRANGE", Args: "key offset value", Version: "2.2.0", Desc: "Overwrite part of a string at key starting at the specified offset"},
	{Command: "SHUTDOWN", Args: "[NOSAVE|SAVE]", Version: "1.0.0", Desc: "Synchronously save the dataset to disk and then shut down the server"},
	{Command: "SINTER", Args: "key [key ...]", Version: "1.0.0", Desc: "Intersect multiple sets"},
	{Command: "SINTERSTORE", Args: "destination key [key ...]", Version: "1.0.0", Desc: "Intersect multiple sets and store the resulting set in a key"},
	{Command: "SISMEMBER", Args: "key member", Version: "1.0.0", Desc: "Determine if a given value is a member of a set"},
	{Command: "SLAVEOF", Args: "host port", Version: "1.0.0", Desc: "Make the server a replica of another instance, or promote it as master. Deprecated starting with Redis 5. Use REPLICAOF instead."},
	{Command: "SLOWLOG", Args: "subcommand [argument]", Version: "2.2.12", Desc: "Manages the Redis slow queries log"},
	{Command: "SMEMBERS", Args: "key", Version: "1.0.0", Desc: "Get all the members in a set"},
	{Command: "SMOVE", Args: "source destination member", Version: "1.0.0", Desc: "Move a member from one set to another"},
	{Command: "SORT", Args: "key [BY pattern] [LIMIT offset count] [GET pattern [GET pattern ...]] [ASC|DESC] [ALPHA] [STORE destination]", Version: "1.0.0", Desc: "Sort the elements in a list, set or sorted set"},
	{Command: "SPOP", Args: "key [count]", Version: "1.0.0", Desc: "Remove and return one or multiple random members from a set"},
	{Command: "SRANDMEMBER", Args: "key [count]", Version: "1.0.0", Desc: "Get one or multiple random members from a set"},
	{Command: "SREM", Args: "key member [member ...]", Version: "1.0.0", Desc: "Remove one or more members from a set"},
	{Command: "SSCAN", Args: "key cursor [MATCH pattern] [COUNT count]", Version: "2.8.0", Desc: "Incrementally iterate Set elements"},
	{Command: "STRLEN", Args: "key", Version: "2.2.0", Desc: "Get the length of the value stored in a key"},
	{Command: "SUBSCRIBE", Args: "channel [channel ...]", Version: "2.0.0", Desc: "Listen for messages published to the given channels"},
	{Command: "SUNION", Args: "key [key ...]", Version: "1.0.0", Desc: "Add multiple sets"},
	{Command: "SUNIONSTORE", Args: "destination key [key ...]", Version: "1.0.0", Desc: "Add multiple sets and store the resulting set in a key"},
	{Command: "SWAPDB", Args: "index index", Version: "4.0.0", Desc: "Swaps two Redis databases"},
	{Command: "SYNC", Args: "-", Version: "1.0.0", Desc: "Internal command used for replication"},
	{Command: "TIME", Args: "-", Version: "2.6.0", Desc: "Return the current server time"},
	{Command: "TOUCH", Args: "key [key ...]", Version: "3.2.1", Desc: "Alters the last access time of a key(s). Returns the number of existing keys specified."},
	{Command: "TTL", Args: "key", Version: "1.0.0", Desc: "Get the time to live for a key"},
	{Command: "TYPE", Args: "key", Version: "1.0.0", Desc: "Determine the type stored at key"},
	{Command: "UNLINK", Args: "key [key ...]", Version: "4.0.0", Desc: "Delete a key asynchronously in another thread. Otherwise it is just as DEL, but non blocking."},
	{Command: "UNSUBSCRIBE", Args: "[channel [channel ...]]", Version: "2.0.0", Desc: "Stop listening for messages posted to the given channels"},
	{Command: "UNWATCH", Args: "-", Version: "2.2.0", Desc: "Forget about all watched keys"},
	{Command: "WAIT", Args: "numreplicas timeout", Version: "3.0.0", Desc: "Wait for the synchronous replication of all the write commands sent in the context of the current connection"},
	{Command: "WATCH", Args: "key [key ...]", Version: "2.2.0", Desc: "Watch the given keys to determine execution of the MULTI/EXEC block"},
	{Command: "XACK", Args: "key group ID [ID ...]", Version: "5.0.0", Desc: "Marks a pending message as correctly processed, effectively removing it from the pending entries list of the consumer group. Return value of the command is the number of messages successfully acknowledged, that is, the IDs we were actually able to resolve in the PEL."},
	{Command: "XADD", Args: "key ID field string [field string ...]", Version: "5.0.0", Desc: "Appends a new entry to a stream"},
	{Command: "XCLAIM", Args: "key group consumer min-idle-time ID [ID ...] [IDLE ms] [TIME ms-unix-time] [RETRYCOUNT count] [force] [justid]", Version: "5.0.0", Desc: "Changes (or acquires) ownership of a message in a consumer group, as if the message was delivered to the specified consumer."},
	{Command: "XDEL", Args: "key ID [ID ...]", Version: "5.0.0", Desc: "Removes the specified entries from the stream. Returns the number of items actually deleted, that may be different from the number of IDs passed in case certain IDs do not exist."},
	{Command: "XGROUP", Args: "[CREATE key groupname id-or-$] [SETID key id-or-$] [DESTROY key groupname] [DELCONSUMER key groupname consumername]", Version: "5.0.0", Desc: "Create, destroy, and manage consumer groups."},
	{Command: "XINFO", Args: "[CONSUMERS key groupname] [GROUPS key] [STREAM key] [HELP]", Version: "5.0.0", Desc: "Get information on streams and consumer groups"},
	{Command: "XLEN", Args: "key", Version: "5.0.0", Desc: "Return the number of entires in a stream"},
	{Command: "XPENDING", Args: "key group [start end count] [consumer]", Version: "5.0.0", Desc: "Return information and entries from a stream consumer group pending entries list, that are messages fetched but never acknowledged."},
	{Command: "XRANGE", Args: "key start end [COUNT count]", Version: "5.0.0", Desc: "Return a range of elements in a stream, with IDs matching the specified IDs interval"},
	{Command: "XREAD", Args: "[COUNT count] [BLOCK milliseconds] STREAMS key [key ...] ID [ID ...]", Version: "5.0.0", Desc: "Return never seen elements in multiple streams, with IDs greater than the ones reported by the caller for each stream. Can block."},
	{Command: "XREADGROUP", Args: "GROUP group consumer [COUNT count] [BLOCK milliseconds] STREAMS key [key ...] ID [ID ...]", Version: "5.0.0", Desc: "Return new entries from a stream using a consumer group, or access the history of the pending entries for a given consumer. Can block."},
	{Command: "XREVRANGE", Args: "key end start [COUNT count]", Version: "5.0.0", Desc: "Return a range of elements in a stream, with IDs matching the specified IDs interval, in reverse order (from greater to smaller IDs) compared to XRANGE"},
	{Command: "XTRIM", Args: "key MAXLEN [~] count", Version: "5.0.0", Desc: "Trims the stream to (approximately if '~' is passed) a certain size"},
	{Command: "ZADD", Args: "key [NX|XX] [CH] [INCR] score member [score member ...]", Version: "1.2.0", Desc: "Add one or more members to a sorted set, or update its score if it already exists"},
	{Command: "ZCARD", Args: "key", Version: "1.2.0", Desc: "Get the number of members in a sorted set"},
	{Command: "ZCOUNT", Args: "key min max", Version: "2.0.0", Desc: "Count the members in a sorted set with scores within the given values"},
	{Command: "ZINCRBY", Args: "key increment member", Version: "1.2.0", Desc: "Increment the score of a member in a sorted set"},
	{Command: "ZINTERSTORE", Args: "destination numkeys key [key ...] [WEIGHTS weight] [AGGREGATE SUM|MIN|MAX]", Version: "2.0.0", Desc: "Intersect multiple sorted sets and store the resulting sorted set in a new key"},
	{Command: "ZLEXCOUNT", Args: "key min max", Version: "2.8.9", Desc: "Count the number of members in a sorted set between a given lexicographical range"},
	{Command: "ZPOPMAX", Args: "key [count]", Version: "5.0.0", Desc: "Remove and return members with the highest scores in a sorted set"},
	{Command: "ZPOPMIN", Args: "key [count]", Version: "5.0.0", Desc: "Remove and return members with the lowest scores in a sorted set"},
	{Command: "ZRANGE", Args: "key start stop [WITHSCORES]", Version: "1.2.0", Desc: "Return a range of members in a sorted set, by index"},
	{Command: "ZRANGEBYLEX", Args: "key min max [LIMIT offset count]", Version: "2.8.9", Desc: "Return a range of members in a sorted set, by lexicographical range"},
	{Command: "ZRANGEBYSCORE", Args: "key min max [WITHSCORES] [LIMIT offset count]", Version: "1.0.5", Desc: "Return a range of members in a sorted set, by score"},
	{Command: "ZRANK", Args: "key member", Version: "2.0.0", Desc: "Determine the index of a member in a sorted set"},
	{Command: "ZREM", Args: "key member [member ...]", Version: "1.2.0", Desc: "Remove one or more members from a sorted set"},
	{Command: "ZREMRANGEBYLEX", Args: "key min max", Version: "2.8.9", Desc: "Remove all members in a sorted set between the given lexicographical range"},
	{Command: "ZREMRANGEBYRANK", Args: "key start stop", Version: "2.0.0", Desc: "Remove all members in a sorted set within the given indexes"},
	{Command: "ZREMRANGEBYSCORE", Args: "key min max", Version: "1.2.0", Desc: "Remove all members in a sorted set within the given scores"},
	{Command: "ZREVRANGE", Args: "key start stop [WITHSCORES]", Version: "1.2.0", Desc: "Return a range of members in a sorted set, by index, with scores ordered from high to low"},
	{Command: "ZREVRANGEBYLEX", Args: "key max min [LIMIT offset count]", Version: "2.8.9", Desc: "Return a range of members in a sorted set, by lexicographical range, ordered from higher to lower strings."},
	{Command: "ZREVRANGEBYSCORE", Args: "key max min [WITHSCORES] [LIMIT offset count]", Version: "2.2.0", Desc: "Return a range of members in a sorted set, by score, with scores ordered from high to low"},
	{Command: "ZREVRANK", Args: "key member", Version: "2.0.0", Desc: "Determine the index of a member in a sorted set, with scores ordered from high to low"},
	{Command: "ZSCAN", Args: "key cursor [MATCH pattern] [COUNT count]", Version: "2.8.0", Desc: "Incrementally iterate sorted sets elements and associated scores"},
	{Command: "ZSCORE", Args: "key member", Version: "1.2.0", Desc: "Get the score associated with the given member in a sorted set"},
	{Command: "ZUNIONSTORE", Args: "destination numkeys key [key ...] [WEIGHTS weight] [AGGREGATE SUM|MIN|MAX]", Version: "2.0.0", Desc: "Add multiple sorted sets and store the resulting sorted set in a new key"},
}
