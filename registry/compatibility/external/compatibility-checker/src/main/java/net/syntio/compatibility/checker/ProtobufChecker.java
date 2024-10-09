package net.syntio.compatibility.checker;

import io.apicurio.registry.content.ContentHandle;
import io.apicurio.registry.rules.compatibility.CompatibilityLevel;
import io.apicurio.registry.rules.compatibility.ProtobufCompatibilityChecker;

import java.util.List;

public class ProtobufChecker implements CompatibilityChecker {
    @Override
    public boolean testCompatibility(CompatibilityLevel level, List<ContentHandle> history, ContentHandle currentSchema) {
        ProtobufCompatibilityChecker cc =  new ProtobufCompatibilityChecker();
        return cc.testCompatibility(level, history, currentSchema).isCompatible();
    }
}
