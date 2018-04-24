# Roadmap

## Slackbot future plans

- Different commands from slack  
   `> @bot help` lists all available commands  
   `> @bot` as comment when uploading an image will execute the default command (replace faces if found, else photobomb)  
   `> @bot photobomb` or `> @bot bomb` will paste a "body" into the image instead of replacing faces  
   `> @bot success` will photobomb the image with a predefined "success kid (Crisu)" body
- Possibility to chain commands, e.g. first replace faces, then "success kid"
- Create and adapter/interface for the face_replace executable. This would improve code quality and make it possible to
   use other libraries for face replacement
