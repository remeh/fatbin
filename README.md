# Fatbin

## Howto

### Create an archive

```
fatbin -dir /the/directory/to/compress -exe executable
```

Current limitation: executable must be in the root of the compressed directory.

### Start an archive

In the directory where there is the archive.fbin file :

```
fatbin
```

or to run a given archive :

```
fatbin /path/to/an/archive/mehstation.fbin
```

## Future work

 * Gzip the archive.
 * Transmits argv and flags to the underlying app.
 * Ship only one executable file instead of fatbin+archive.
