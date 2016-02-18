# dict
Golang nested maps map[string]interface{}

## Usage
```go

m := dict.Map{
  "name": Map{
    "first": "Morpit",
    "last": "Jonserred",
  },
  "age": 47,
}

m.Get("name.first") // == "Morpit"
m.Get("age") // == 47

m.Set("name.middlename", "Mortimer")


```
