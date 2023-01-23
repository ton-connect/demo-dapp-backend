# demo-dapp-backend

The example of [demo-dapp](https://github.com/ton-connect/demo-dapp-with-backend) backend with authorization by ton address.

Authorization process is:
1. Client fetches payload to be signed by wallet:
```
<host>/ton-proof/generatePayload

response: 
"E5B4ARS6CdOI2b5e1jz0jnS-x-a3DgfNXprrg_3pec0="
```

2. Client connects to the wallet via TonConnect 2.0 and passes `ton_proof` request with specified payload.
See the [frontend SDK](https://github.com/ton-connect/sdk/tree/main/packages/sdk) for more details.

3. User approves connection and client receives signed payload with additional prefixes.
4. Client sends signed result to the backend. Backend checks correctnes of the all prefixes and signature correctness and returns auth token:
```
<host>/ton-proof/checkProof
{
  "address": "0:f63660ff947e5fe6ed4a8f729f1b24ef859497d0483aaa9d9ae48414297c4e1b", // user's address
  "network": "-239", // "-239" for mainnet and "-1" for testnet
  "proof": {
      "timestamp": 1668094767, // unix epoch seconds
    "domain": {
     "lengthBytes": 21,
      "value": "ton-connect.github.io"
    },
    "signature": "28tWSg8RDB3P/iIYupySINq1o3F5xLodndzNFHOtdi16Z+MuII8LAPnHLT3E6WTB27//qY4psU5Rf5/aJaIIAA==",
    "payload": "E5B4ARS6CdOI2b5e1jz0jnS-x-a3DgfNXprrg_3pec0=" // payload from the step 1.
  }
}

response: 
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZGRyZXNzIjoiMDpmNjM2NjBmZjk0N2U1ZmU2ZWQ0YThmNzI5ZjFiMjRlZjg1OTQ5N2QwNDgzYWFhOWQ5YWU0ODQxNDI5N2M0ZTFiIiwiZXhwIjoxNjY4MDk4NDkwfQ.13sg3Mgt2hT9_vChan3bmQkp_Wsigj9YjSoKABTsVGA"
}
```

See `ton_proof` details in the [docs](https://github.com/ton-connect/docs/blob/main/requests-responses.md#address-proof-signature-ton_proof).

5. Client can access auth-required endpoints:
```
<host>/dapp/getAccountInfo?network=-239
Bearer <token>

response:
json
```

# Ton Proof JS verification
You can find an example of the ton_proof verification in JavaScript [here](https://gist.github.com/TrueCarry/cac00bfae051f7028085aa018c2a05c6).
