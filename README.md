This is a POC for a compression algorithm.

The logic behind it was told to me by my cats in my dreams, and I don't care how efficient it is, please leave me alone.

We need to be careful with compressed buffers, segments and segment positions, since coordinates needs to represent the final position in the decompressed buffer.

First we remove all repetitions, since they are extremely easy to detect and compact, while adding complexity for other layers.

Then we deduplicate similar groups of repetitions, merging their coordinates.

Then we use repetition groups, in which we detect the same pattern repeating many times, and we try to grow these groups as large
as possible.

## Data structure in disk

### Block

|Address|Size|Name|Description|
|--- |---| ---|---|
|0x00|4 bytes|Size|Size of the original buffer|

### Ordered Segment

|Address|Size|Name|Description|
|--- |---| ---|---|
|0x00|1 byte|Metadata|Stores segment-type, repeat-size,order-len|
|+1|4 bytes|Buffer length|4294967295 bytes, or 4.294967 gigabytes|
|+5|1~2 bytes|RepeatCount|How many times buffer is repeated|
|+1 or +2|1*LEN bytes|Order|The order in which the segments should decompress|
|+len(Order)|1*LEN bytes|Buffer|Buffer containing compressed data|

So ideally we have a 6 bytes + 1 * LEN overhead, meaning sections that save more than 7 bytes
are compressing the file.

### Ordered Segment Metadata


|Bit|Description|
|--- | ---|
|1-2|Segment Type|
|3|Repeat Size: 0 = 1 byte, 1 = 2 bytes|
|4,5,6,7,8|len(Order), Max 32|