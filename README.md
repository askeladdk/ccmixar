# C&C Mix Archiver

CCMIXAR is a command line tool for managing .mix archive files that are used by several C&C games. It can pack and unpack checksummed and encrypted .mix archives for C&C Tiberian Dawn, Tiberian Sun, Red Alert 1 and Red Alert 2. Compared to OmniBlade's ccmix, CCMIXAR is more efficient (faster) and does not contain any crufty C++ code written over a decade ago.

## Usage

### Pack a directory in a .mix file

`ccmixar pack -game <cc1|cc2|ra1|ra2> -mix <outpath> -dir <inpath> [-checksum] [-database] [-encrypt]`

### List content information of .mix file

`ccmixar info -game <cc1|cc2|ra1|ra2> -mix <inpath>`

### Unpack a .mix file to a directory

`ccmixar unpack -game <cc1|cc2|ra1|ra2> -mix <inpath> -dir <outpath>`

## Acknowledgements

OmniBlade for his work reverse engineering the .mix file encryption algorithm and writing his ccmix tool which ccmixar is inspired by.

Olaf van der Spek for his work reverse engineering the C&C files formats and the tools he has developed over the years to allow modding these classic games.
