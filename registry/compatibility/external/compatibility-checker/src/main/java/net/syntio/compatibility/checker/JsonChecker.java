package net.syntio.compatibility.checker;

import io.apicurio.registry.content.ContentHandle;
import io.apicurio.registry.rules.compatibility.CompatibilityLevel;
import io.apicurio.registry.rules.compatibility.JsonSchemaCompatibilityChecker;

import java.util.List;

public class JsonChecker implements CompatibilityChecker {
    @Override
    public boolean testCompatibility(CompatibilityLevel level, List<ContentHandle> history, ContentHandle currentSchema) {
        JsonSchemaCompatibilityChecker cc =  new JsonSchemaCompatibilityChecker();
        return cc.testCompatibility(level, history, currentSchema).isCompatible();
    }
}
