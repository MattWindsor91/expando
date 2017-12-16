# Expando

`expando` generates sentences by nondeterministically expanding macros.

## Example

```
$ ./expando -n 10 example_macros/sentences
Debate-Jivingly Go Live, Musician!
Slowly Get Up, Old Drama!
Rapidly Go West, Incredible Student!
Slow Down Fondly, Darts!
Slow Down Amazingly, Silence!
Radio-tacular Song Played Stupendously
Get Straight Violently, Stupefyingly Hairy Song!
The Melodic Knives Wonderfully Enqueued
The Insidious Radio Fondly Broadcasts
Go West Violently, Stupendously Green Projects!
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
