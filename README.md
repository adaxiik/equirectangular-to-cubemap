# equirectangular-to-cubemap
Convert equirectangular image to cubemap (6 faces stored separately)

# Build
```sh
$ go build ./e2c.go
```

# Usage
```sh
$ ./e2c <output_size> <input_image> <output_folder>
```
Example:
```sh
$ ./e2c 512 skybox.png output
```

Or directly:
```sh
$ go run ./e2c.go 512 skybox.png output
```
# Notes
- It's able to load jpg and png images, but output is always png. (although it can be changed easily)

# Go?
- I just wanted to try it.. this is literally my first program in Go. :)