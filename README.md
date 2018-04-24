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
    $ go get -u github.com/spf13/viper
    ```
    
3. Build the bot

    ```
    $ cd slack/src/faceswapbot
    $ go build
    ```

4. Create the `config.yml` file

    ```yaml
    slack:
      token: BOT-TOKEN
    
    pyswapper:
      paths:
        executable: /path/to/face_replace/face_replace.py
        faces:      /path/to/faces/*.png
        bodies:     /path/to/bodies/*.png
        success:    /path/to/success.png OR /path/to/success/*.png
    ```

5. Run the bot

    ```
    /path/to/faceswapbot -config /path/to/config.yml
    ```