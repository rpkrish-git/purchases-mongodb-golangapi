# purchases-mongodb-golangapi
This is my Multi Currency Purchase Transaction API with Golang &amp; MongoDB

*Pre-Requisites:*

Docker should be started and running

MongoDB should be started and running on the port: 27017

(External API) Treasury Reporting Rates of Exchange API must be available and running

*Purchase API has following end points:*

healthcheck (GET): http://localhost:6000/health

create a transaction (POST): http://localhost:6000/purchase
    input body: ```{
    "description" : "Brother Laser Printer",
    "transactionDateTime" : "2020-06-20T15:02:05Z",
    "purchaseAmount" : 350.00
}```

update a transaction (PUT): http://localhost:6000/purchases/652dee955c34520ecad8c491
    input body: ```{
    "description" : "keyboard-mouse-combo",
    "transactionDateTime" : "2023-10-20T15:04:05Z",
    "purchaseAmount" : 52.35
}```

delete a transaction (DELETE): http://localhost:6000/purchases/652dee955c34520ecad8c492

retrieve a transaction (POST): http://localhost:6000/purchases/652dee955c34520ecad8c492
    input body: ```{"currency":"australia-dollar"}```

retrieve all transactions (POST): http://localhost:6000/purchases   
    input body: ```{"currency":"australia-dollar"}```

