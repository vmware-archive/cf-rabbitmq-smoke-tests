package io.pivotal.pcf.rabbitmq.springexampleapp;


import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.io.IOException;
import java.util.concurrent.TimeoutException;

import static io.pivotal.pcf.rabbitmq.springexampleapp.SpringExampleApp.QUEUE_NAME;
import static org.springframework.web.bind.annotation.RequestMethod.GET;
import static org.springframework.web.bind.annotation.RequestMethod.POST;

@RestController
@RequestMapping(value = "/queues/" + QUEUE_NAME)
public class QueuesController {

    private final RabbitTemplate rabbitTemplate;

    public QueuesController(RabbitTemplate rabbitTemplate) {
        this.rabbitTemplate = rabbitTemplate;
    }

    @RequestMapping(method = POST, consumes = {"text/plain"})
    public void publishMessage(@RequestBody String message) throws IOException, TimeoutException {
        rabbitTemplate.convertAndSend(QUEUE_NAME, QUEUE_NAME, message);
    }

    @RequestMapping(method = GET, produces = {"text/plain"})
    public String getMessages() {
        return new String(rabbitTemplate.receive(QUEUE_NAME).getBody());
    }
}
