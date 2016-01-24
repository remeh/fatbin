# Fatbin

Instead of shipping a ZIP containing resources (images, sounds, etc.) and an executable, `fatbin` permits to compress everything in an unique executable file.

It's my entry to the GopherGala 2016.

## Howto

Example video : https://c.remy.io/oweinxoW

### Create an archive

Usage:

```
-f.dir string
        the directory to fatbinerize
-f.exe string
        the file inside the fatbin archive to execute at startup
-f.out string
        the archive file to create. (default "archive.fbin")
```

Example:

```
fatbin -f.dir /path/to/my/program -f.exe main -f.out program.fbin
```

Current limitation: the executable runned at startup must in the root of the compressed directory.

### Start an archive

The archive is an executable which can directly be started.

```
./program.fbin
```

## Future work

 * Compress the archive with something else than gzip.
 * Ship "surrounding" files in the temporary execution context. (e.g. for .conf)

## License

fatbin, created by Rémy 'remeh' Mathieu, is under the terms of the MIT License.
