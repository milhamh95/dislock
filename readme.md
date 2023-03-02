# Read me

- concurrent test with vegeta, for mac -> `brew install vegeta`

```
echo "GET http://localhost:1323/counter2" | vegeta attack -duration=30s -rate=10 | vegeta report --type=text
```
