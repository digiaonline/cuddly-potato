# cuddly-potato

[![Build Status](https://travis-ci.org/digiaonline/cuddly-potato.svg?branch=master)](https://travis-ci.org/digiaonline/cuddly-potato)

Hammertime!

## Setup

### Python stuff

1. Install Python 3.6 (or later)
2. Install OpenCV:

   ```
   pip3 install opencv-python
   ```


### Go stuff
1. Install Go

    ```
    https://golang.org/doc/install
    ```

2. Install [Slack API in Go](https://github.com/nlopes/slack)

    ```
    $ go get -u github.com/nlopes/slack
    ```
    
3. Build the bot

    ```
    $ cd slack/src/faceswapbot
    $ go build
    ```

4. Run the bot

    ```
    ./faceswapbot -token [BOT_TOKEN] -faceSwapper [/path/to/face_replace.py]
    ```