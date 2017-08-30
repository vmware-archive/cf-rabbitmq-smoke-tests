package io.pivotal.pcf.rabbitmq.springexampleapp;


import org.springframework.web.bind.annotation.*;

import static org.springframework.web.bind.annotation.RequestMethod.*;

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
}
