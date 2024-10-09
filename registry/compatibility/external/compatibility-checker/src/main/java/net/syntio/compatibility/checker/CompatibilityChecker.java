package net.syntio.compatibility.checker;

import io.apicurio.registry.content.ContentHandle;
import io.apicurio.registry.rules.compatibility.CompatibilityLevel;

import java.util.List;

public interface CompatibilityChecker {
    boolean testCompatibility(CompatibilityLevel level, List<ContentHandle> history, ContentHandle currentSchema);
}
