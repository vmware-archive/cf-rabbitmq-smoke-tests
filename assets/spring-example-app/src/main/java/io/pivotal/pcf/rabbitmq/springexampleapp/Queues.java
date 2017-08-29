package io.pivotal.pcf.rabbitmq.springexampleapp;


import org.springframework.web.bind.annotation.*;

import static org.springframework.web.bind.annotation.RequestMethod.*;

@RestController
@RequestMapping(value = "/queues")
public class Queues {
    @RequestMapping(value = "/{queueName}", method = POST, consumes = {"text/plain"})
    public String publishMessage(@PathVariable("queueName") String queueName, @RequestBody String message) {
        return queueName + message;
    }
}
