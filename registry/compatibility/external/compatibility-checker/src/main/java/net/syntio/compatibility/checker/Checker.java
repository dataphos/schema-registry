package net.syntio.compatibility.checker;

import io.apicurio.registry.content.ContentHandle;
import io.apicurio.registry.rules.compatibility.CompatibilityLevel;
import net.syntio.compatibility.Message;
import net.syntio.compatibility.CheckerFactory;

import java.util.ArrayList;
import java.util.List;

public class Checker {
    public static boolean checkCompatibility(Message msg, List<String> history, CompatibilityLevel mode) throws Exception {
        ContentHandle schema = ContentHandle.create(msg.getSchema());
        List<ContentHandle> schemaHistory = new ArrayList<>();

        for (String s : history) {
            ContentHandle ps = ContentHandle.create(s);
            schemaHistory.add(ps);
        }
        CompatibilityChecker cc = CheckerFactory.createChecker(msg.getFormat());

        return cc.testCompatibility(mode, schemaHistory, schema);
    }
}
