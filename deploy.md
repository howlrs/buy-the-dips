```sh gcloud to schedules
gcloud scheduler jobs create http \
    --name=buy-the-dips-each-every-minute \
    --schedule="* * * * *" \
    --uri="api endpoint: https://example.com/your-endpoint" \
    --http-method=GET
```

```sh gcloud to run
gcloud run deploy btd --source .
```