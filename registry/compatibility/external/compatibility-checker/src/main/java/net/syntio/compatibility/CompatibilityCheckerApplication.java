package net.syntio.compatibility;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

import java.util.Collections;

@SpringBootApplication
public class CompatibilityCheckerApplication {
    public static void main(String[] args) {
        SpringApplication app = new SpringApplication(CompatibilityCheckerApplication.class);
        app.setDefaultProperties(Collections.singletonMap("server.port", "8088"));
        app.run(args);
    }
}
