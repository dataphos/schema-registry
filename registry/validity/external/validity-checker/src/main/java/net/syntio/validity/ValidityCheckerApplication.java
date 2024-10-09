package net.syntio.validity;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

import java.util.Collections;

@SpringBootApplication
public class ValidityCheckerApplication {
    public static void main(String[] args) {
        SpringApplication app = new SpringApplication(ValidityCheckerApplication.class);
        app.setDefaultProperties(Collections.singletonMap("server.port", "8089"));
        app.run(args);
    }

}
