This is a POC for a compression algorithm.

The logic behind it was told to me by my cats in my dreams, and I don't care how efficient it is, please leave me alone.

We need to be careful with compressed buffers, segments and segment positions.

For each compression logic, you should create a layer and order of execution, to prevent the coordinates of one layer from leaking.

First we remove all repetitions, since they are extremely easy to detect and compact, while adding complexity for other layers.

Then we use repetition groups, in which we detect the same pattern repeating many times, and we try to grow these groups as large
as possible.