package net.syntio.validity.checker;

import io.apicurio.registry.content.ContentHandle;
import io.apicurio.registry.rules.RuleViolationException;
import io.apicurio.registry.rules.validity.ContentValidator;
import io.apicurio.registry.rules.validity.ValidityLevel;

import net.syntio.validity.ValidatorFactory;

import java.util.Collections;

public class Checker {
    public static boolean checkValidity(String schemaType, String schema, String mode) {
        ValidityLevel valLevel = switch (mode.toLowerCase()) {
            case "syntax-only" -> ValidityLevel.SYNTAX_ONLY;
            case "full" -> ValidityLevel.FULL;
            default -> ValidityLevel.NONE;
        };

        ContentValidator validator = ValidatorFactory.createValidator(schemaType);
        ContentHandle contentHandle = ContentHandle.create(schema);
        try {
            validator.validate(valLevel, contentHandle, Collections.emptyMap());
            return true;
        } catch (RuleViolationException e) {
            return false;
        }
    }

}
