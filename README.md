# Expando

`expando` generates sentences by nondeterministically expanding macros.

## Example

```
$ ./expando -n 10 example_macros/sentences
Legume-Entranceingly Go West, Articulate Idea!
Wonderfully Slow Down, Insidious Melodies!
Wonderfully Bodacious Discourse Harmed Sarcastically
Entertainment Scheduled Stupefyingly
Mysterious Trickster Harmed
Speed Up, Tuneful Sandwich!
Get Straight, Face-Pleaseingly Incredible Face!
Boringly Unbelievable Miscellanea Break Stupefyingly
Colourless Managers Enjoyed Boringly
The Slowly Harmonic Knives Regarded Stupendously
```

## Usage

```
expando [-n TIMES] [FILE]
```

- `TIMES` is the number of times to expand the base macro (default: 1)
- `FILE` is the macro file to load; `expando` reads from stdin if it is missing.

To read in multiple macro files, use, for example:

```
cat FILE1 FILE2 | expando
```

First, of course, make sure `FILE1` ends with `%%`, otherwise the macros from
`FILE1` might crash into those in `FILE2`!

## Macro file format

See `example_macros/sentences`: that file contains comments describing the format,
and an example of the format at work.

## Licence

MIT; see `LICENSE`.
