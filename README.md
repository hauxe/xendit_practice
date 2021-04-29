# Marvel Character Service

This service use [Marvel API](https://developer.marvel.com) to build a service to query character list and character info

## Installation

Install Golang at official [page](https://golang.org/dl)

This service requires >=Go 1.5

This require Go Module annd it will be automatically download into your machine in the first build

```bash
go build ./cmd/marvel -o marvel
```

Start service by running the binary. The service requires Marvel API Public Key and API Private Key.
Both keys must be passed in the service by 2 Environment Variable: API_PUBLIC_KEY and API_PRIVATE_KEY

```bash
API_PUBLIC_KEY={public_key} API_PRIVATE_KEY={private_key} ./marvel
```

The service always listen on localhost port 8080

## Usage

Use HTTP Rest API provided below to access Service

/characters
Get list of all Marvel's characters id

/characters/{character_id}
Get character innformation (ID, Name, Description) of a character by id

## License

[MIT](https://choosealicense.com/licenses/mit/)
