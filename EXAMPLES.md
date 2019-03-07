# Examples
In gofigure, there are a number of features designed to make your life easier.

One of these features is the expansion macro. You can effectively expand roots as children of the root structure, or any other keymap.
```
[root.%{root1,root2}] 
# The statement above will prefix everything below this up until the file ends 
# or until the section is left with the "[]" operator, with root.root1. and root.root2. 
# (creating a copy for every root within the expansion macro)

key:"value"

[] # leave the current section

rootKey: "rootkey"
```
This effectively expands the keys `root1` and `root2` as children of `root` and prefixes everything in that section with those keys as parents. In JSON, it looks like this
```
{
  "root": {
    "root1": {
      "key": "value"
    },
    "root2": {
      "key": "value"
    }
  },
  "rootKey": "rootkey"
}
```
Now, if we want to select all children of a parent we can to so by using the `@` operator. The `@` operator is a selector which selects all predefined maps at the given level.
Say for example we wanted to add something to both root1 and root2 in the above example. We would add to the end of the file
```
[root.@]
added: "more value"
```
Both `root1` and `root2` along with any other keys defined as children of `root`, will now contain the added value. For sake of completeness, the JSON will now look like this
```
{
  "root": {
    "root1": {
      "key": "value",
      "added": "more value"
    },
    "root2": {
      "key": "value"
      "added": "more value"
    }
  },
  "rootKey": "rootkey"
}
```