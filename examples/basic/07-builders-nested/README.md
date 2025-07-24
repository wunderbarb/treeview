# Nested Builder

This example demonstrates how to transform custom nested organizational data types into tree
nodes using the nested builder. It walks you through converting your existing hierarchical data
structures (like departments with sub-departments) into renderable trees, applying conditional
expansion to control which nodes start open, and using custom presentation providers instead of
building off the default provider. This is ideal when you already have nested data and just
want to visualize it.

## Running the Example

```bash
go run main.go
```

## Demo

![The Demo](https://vhs.charm.sh/vhs-7zo22mFV7SOdiyVCQhemof.gif)
