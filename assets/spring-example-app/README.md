```
gradle build
cf create-service p-rabbitmq standard rabbitmq
cf create-security-group rabbitmq asg-rabbitmq.yml
cf bind-security-group rabbitmq system system
cf push
curl -X POST http://pivotal-rabbitmq.bosh-lite.com/queues/<queueName> -d "the message contents" -H "content-type: text/plain"
curl -X GET http://pivotal-rabbitmq.bosh-lite.com/queues/<queueName>
```
