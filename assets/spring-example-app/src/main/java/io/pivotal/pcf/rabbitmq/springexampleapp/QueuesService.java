package io.pivotal.pcf.rabbitmq.springexampleapp;

import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Component;

@Component
public class QueuesService {
    private final RabbitTemplate rabbitTemplate;

    public QueuesService(RabbitTemplate rabbitTemplate) {
        this.rabbitTemplate = rabbitTemplate;
    }

    public void publish(String queueName, String message) {
        rabbitTemplate.convertAndSend(queueName, message);
    }
}
