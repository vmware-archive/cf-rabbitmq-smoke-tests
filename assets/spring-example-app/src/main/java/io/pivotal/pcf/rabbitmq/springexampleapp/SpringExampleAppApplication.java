package io.pivotal.pcf.rabbitmq.springexampleapp;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.stereotype.Component;

@SpringBootApplication
public class SpringExampleAppApplication {

	public static void main(String[] args) {
 		SpringApplication.run(SpringExampleAppApplication.class, args);
	}
}
