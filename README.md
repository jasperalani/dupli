# dupli
Command line duplicate image finder written in Go.

Dupli will find your duplicates, copy them into a folder, and delete the original duplicate.

Dupli only compares png, jpeg, jpg

Dupli does not compare images with different dimensions.

Dupli uses [imgdiff](https://github.com/n7olkachev/imgdiff) to work.

### Usage
`./dupli -scan`<br>
`./dupli -scan -loc=/home/dupli/pictures`

### Todo
- Create app folder (somewhere) for log storage
- Scan subfolders


##### Version  1.0.1