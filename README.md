# legoizer

**Legoizer** is a simple tool to generate lego-like images taking as input the original image and converting it to lego bricks. This process consists in two steps: 
1. The image is converted to it's quantified representation to reduce the number of colors. 
2. The bricks are generated based on the curently processed cell color compared to it's neighboring cells average color.

### Install

`$ go get -u github.com/esimov/legoizer`

### Run
Type `legoizer --help` to get the currently supported commands.

```
Usage of legoizer:
  -colors int
    	Number of colors (default 128)
  -in string
    	Input path
  -out string
    	Output path
  -size int
    	Lego size     
```     

The application take the source image and output the generated image into the destination folder.

| Source image | Legoized image
|:--:|:--:|
| <img src="https://user-images.githubusercontent.com/883386/27582530-e4095cd8-5b39-11e7-97f4-1a457857c80c.png"> | <img src="https://user-images.githubusercontent.com/883386/27582636-42c42d84-5b3a-11e7-8f60-15ca7cf4f2ce.png"> |
| <img src="https://user-images.githubusercontent.com/883386/27582916-54d7b27e-5b3b-11e7-84b7-5209b878c2ca.jpg" > | <img src="https://user-images.githubusercontent.com/883386/27582932-67795126-5b3b-11e7-82bd-4c4df11d4f5a.png"> |
| <img src="https://user-images.githubusercontent.com/883386/27582571-fea42c9e-5b39-11e7-8357-6ed2a425fdd1.jpg"> | <img src="https://user-images.githubusercontent.com/883386/27582651-4d1cb99a-5b3a-11e7-8bd1-1095d265b373.png"> |  

### Discalimer

This is a simple toy, so it does not have any commercial usage. However if exists any interest for this app (like the possibility to count the needed type of bricks for the construction of a possible real lego), or to create a web interface i will be happy to develop it further.

## License

This software is distributed under the MIT license found in the LICENSE file.

