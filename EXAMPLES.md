# Examples
In gofigure, there are a number of features designed to make your life easier.

### Expansion macros
One of these features is the expansion macro. You can effectively expand roots as children of the root structure, or any other keymap.
```
# Can use both comma and space here
#            |      |
#            V      V
[root.%{root1,root2} root3]
# The statement above will prefix everything below this up until the file ends 
# or until the section is left with the "[]" operator, with root.root1., root.root2. and root3. 
# (creating a copy for every root within the expansion macro)

key: [
  "value",  ; comma if you want
  "value2"  # or don't
  "value3"  
]

; This is also a comment

[] # leave the current section

rootKey: "rootkey"
```
This effectively expands the keys `root1` and `root2` as children of `root`, and `root3` as a separate root. Then it prefixes everything in that section with those keys as parents. In JSON, it looks like this
```
{
  "root": {
    "root1": {
      "key": ["value", "value2", "value3"]
    },
    "root2": {
      "key": ["value", "value2", "value3"]
    }
  },
  "root3": {
    "key": ["value", "value2", "value3"]
  },
  "rootKey": "rootkey"
}
```
### The @ selector
Now, if we want to select all children of a parent we can to so by using the `@` operator. The `@` operator is a selector which selects all predefined maps at the given level.
Say for example we wanted to add something to both root1 and root2 in the above example. We would add to the end of the file
```
[root.@]
added: "more value"
```
Both `root1` and `root2` along with any other keys (which may be) defined as children of `root`, will now contain the added value. For sake of completeness, the JSON will now look like this
```
{
  "root": {
    "root1": {
      "added": "more value",
      "key": ["value", "value2", "value3"]
    },
    "root2": {
      "added": "more value",
      "key": ["value", "value2", "value3"]
    }
  },
  "root3": {
    "key": ["value", "value2", "value3"]
  },
  "rootKey": "rootkey"
}
```
