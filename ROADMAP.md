# Roadmap

## Slackbot future plans

- Possibility to chain commands, e.g. first replace faces, then "success kid"
- Create and adapter/interface for the face_replace executable. This would improve code quality and make it possible to
   use other libraries for face replacement

## Split the code
Slackbot and the python face replacer should both reside in their own repositories. This will help to keep the code
clean and that way [dep](https://github.com/golang/dep) for go could be utilized.
