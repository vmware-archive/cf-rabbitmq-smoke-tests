package io.pivotal.pcf.rabbitmq.springexampleapp;


import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import static org.springframework.web.bind.annotation.RequestMethod.GET;
import static org.springframework.web.bind.annotation.RequestMethod.POST;

@RestController
@RequestMapping(value = "/queues")
public class QueuesController {

    private QueuesService queues;

    public QueuesController(QueuesService queues) {
        this.queues = queues;
    }

    @RequestMapping(value = "/{queueName}", method = POST, consumes = {"text/plain"})
    public void publishMessage(@PathVariable("queueName") String queueName, @RequestBody String message) {
        queues.publish(queueName, message);
    }

    @RequestMapping(value = "/{queueName}", method = GET, produces = {"text/plain"})
    public String getMessages(@PathVariable("queueName") String queueName) {
        return queues.consume(queueName);
    }
}
