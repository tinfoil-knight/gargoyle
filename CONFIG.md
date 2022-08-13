# Configuration

- Supported Format: JSON
- Each config file has a list of services.
- Each service has it's own separate configuration.
- Multiple services can't use the same `source` value.

## Examples

### A Reverse Proxy
```json
[
	{
		// TCP port to listen on, required
		"source": ":8082",
		"reverse_proxy": {
			// url(s) to proxy request to, non-empty required
			"targets": ["http://localhost:3030"]
		}
	}
]
```

### Reverse Proxy w/ Load Balancing & Health-Checks 
```json
[
	{
		"source": ":8080",
		"reverse_proxy": {
			"targets": [
				"http://localhost:3030",
				"http://localhost:3040",
				"http://localhost:3050"
			],
			// load balancing algorithm to use, allowed: "random", "round-robin", default: "random"
			"lb_algorithm": "round-robin",
			"health_check": {
				"enabled": true,
				// relative path for healthcheck requests, default: ""
				"path": "/health",
				// time after which each healthcheck starts, required, unit: seconds
				"interval": 10,
				// HTTP request timeout for requests to target servers, default: 5, unit: seconds
				"timeout": 15
			}
		}
	}
]
```

### File Server
```json
[
	{
		"source": ":8091",
		"fs": {
			// location of directory to serve
			"path": "/Users/tinfoil-knight/Desktop/my-first-site"
		}
	}
]
```
- Static site will be served if an `index.html` file is present at root.

> Note: Config for reverse proxy or file server is skipped & denoted by `// ...` in the examples below.

### Modify Response Headers
```json
[
	{
		// ...
		"header": {
			// map of HTTP headers to add
			"add": {
				"Access-Control-Max-Age": "86400"
			},
			// list of HTTP headers to remove
			"remove": ["Served-By"]
		}
	}
]
```

### Add Authentication

Supported Methods: `basic_auth`, `key_auth`

- Basic HTTP Auth
	```json
	[
		{
			// ...
			"auth": {
				// map of username, hashes
				"basic_auth": {
					"tinfoil": "JDJhJDEwJHB3YWI3YTJPVmxPTG1pTjlaSG5VaU9NM2tUZWZWaTFrSGR4bFg3VXVXTGVpcWkydVA2L2VX",
					"knight": "JDJhJDEwJFB1ZVRaL2dFL1RDS1RxbFc5dTdBYWVEc245OTVuS3FPdGJjeGpXQ3Q5T0RJSjRnT2dEU3lp"
				}
			}
		}
	]
	```
	- Hashes are encoded in base64.
	- Hashing algorithm used is bcrypt.

- API Key Auth
	```json
	[
		{
			// ...
			"auth": {
				"key_auth": {
					// HTTP header to get the key from, default: X-Api-Key
					"header": "X-Auth-Key",
					"key": "some-secret-key"
				}
			}
		}
	]
	```

