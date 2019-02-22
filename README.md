# SecureVault
A stashing utility for storing sensitive data by encrypting it at source, which can be queried easily


**Usage:**
```
stashapp <operation> <arguments>
stashapp [create|publish|fetch|get|add|delete] [-k key | -v value | -c connector | -n name | -p password]
```

- value : escaped json string
- connector: default is filesystem, "google" updates and pulls data from your google drive
- name: Identifier or name for the resource, depends on the connector type. 
- password: All operations need a password to work. Create and Fetch can set the password value for the first time.

**Operations:**
- create : Create the local stash for the first time from a valid raw/un-encrypted JSON. Setting a password is must.
{
  "fruits": {
    "watermelon": [
      "red",
      "green"
    ],
    "banana": "yellow",
    "apple": "red"
  },
  "Vegetables": {
    "spinach": "green",
    "turnip": "white",
    "peas": "green"
  }
}
	stashapp create -n /d/Codethon/sample.json -p password123


- publish: Publishes a local stash to remote location. 
	stashapp publish -n MyEncryptedData.json -p password123 -c google
	stashapp publish -n /d/Codethon/MyEncryptedData.json -p password123

- fetch: Creates a new local stash using an existing encrypted JSON from a remote location.
	stashapp fetch -n /d/Codethon/MyEncryptedData.json -p password
	stashapp fetch -n MyEncryptedData.json -p password123 -c google

- get: Get a key:value pair from local stash. The key paramter can be a valid regular expression. 
	stashapp get -k fruits -p password123
	stashapp get -k uits -p password123

- add: Adds a new key value to the local stash.
	stashapp add -k vegies -v {\"type\":\"greens\"} -p password123

- delete: Deletes an existing key from the local stash.
	stashapp delete -k vegies -p password123
