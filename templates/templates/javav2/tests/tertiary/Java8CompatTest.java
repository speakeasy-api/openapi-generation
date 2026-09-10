package org.openapis.tertiary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.nio.charset.CharacterCodingException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.stream.Collectors;

import org.junit.jupiter.api.Test;
import org.openapis.tertiary.openapi.utils.Java8Compat;

/**
 * Focused units for the java8 arm polyfills. These classes are emitted only on
 * the java8/okhttp arm (isJava8), so the test lives under tertiary. Guards the
 * JDK 9+ semantics the shared templates assume, in particular the fail-fast on
 * duplicate keys/elements that Set.of/Map.of provide.
 */
public class Java8CompatTest {

    @Test
    public void testListOf() {
        List<String> list = Java8Compat.listOf("a", "b", "a");
        assertEquals(3, list.size());
        assertEquals("a", list.get(0));
        assertThrows(UnsupportedOperationException.class, () -> list.add("c"));
    }

    @Test
    public void testSetOfDeduplicatesEagerlyByThrowing() {
        Set<String> set = Java8Compat.setOf("a", "b");
        assertEquals(2, set.size());
        assertThrows(UnsupportedOperationException.class, () -> set.add("c"));
    }

    @Test
    public void testSetOfThrowsOnDuplicate() {
        assertThrows(IllegalArgumentException.class, () -> Java8Compat.setOf("a", "a"));
    }

    @Test
    public void testMapOf() {
        Map<String, Integer> map = Java8Compat.mapOf("a", 1, "b", 2);
        assertEquals(1, map.get("a"));
        assertEquals(2, map.get("b"));
        assertThrows(UnsupportedOperationException.class, () -> map.put("c", 3));
    }

    @Test
    public void testMapOfThrowsOnDuplicateKey() {
        assertThrows(IllegalArgumentException.class, () -> Java8Compat.mapOf("a", 1, "a", 2));
    }

    @Test
    public void testMapOfThrowsOnNullKeyOrValue() {
        assertThrows(NullPointerException.class, () -> Java8Compat.mapOf(null, 1));
        assertThrows(NullPointerException.class, () -> Java8Compat.mapOf("a", null));
    }

    @Test
    public void testSetCopyOf() {
        Set<String> copy = Java8Compat.setCopyOf(Java8Compat.listOf("a", "b"));
        assertEquals(2, copy.size());
        assertThrows(UnsupportedOperationException.class, () -> copy.add("c"));
    }

    @Test
    public void testFailedFutureCompletesExceptionally() {
        RuntimeException cause = new RuntimeException("boom");
        CompletableFuture<String> future = Java8Compat.failedFuture(cause);
        assertTrue(future.isCompletedExceptionally());
        ExecutionException thrown = assertThrows(ExecutionException.class, future::get);
        assertEquals(cause, thrown.getCause());
    }

    @Test
    public void testWriteStringRoundTrip() throws Exception {
        Path path = Files.createTempFile("java8compat", ".txt");
        try {
            String content = "héllo wörld";
            Path returned = Java8Compat.writeString(path, content);
            assertEquals(path, returned);
            assertEquals(content, new String(Files.readAllBytes(path), StandardCharsets.UTF_8));
        } finally {
            Files.deleteIfExists(path);
        }
    }

    @Test
    public void testIsBlank() {
        assertTrue(Java8Compat.isBlank(null));
        assertTrue(Java8Compat.isBlank("  "));
        assertFalse(Java8Compat.isBlank("x"));
        // String.isBlank uses Character.isWhitespace, which covers whitespace
        // above U+0020 (e.g. U+2000 EN QUAD) that String.trim does not
        assertTrue(Java8Compat.isBlank("\u2000\n\t"));
    }

    @Test
    public void testOr() {
        Optional<String> present = Optional.of("a");
        assertEquals(present, Java8Compat.or(present, () -> Optional.of("b")));
        assertEquals(Optional.of("b"), Java8Compat.or(Optional.empty(), () -> Optional.of("b")));
        assertEquals(Optional.empty(), Java8Compat.or(Optional.empty(), Optional::empty));
        assertThrows(NullPointerException.class, () -> Java8Compat.or(Optional.empty(), null));
        assertThrows(NullPointerException.class, () -> Java8Compat.or(Optional.empty(), () -> null));
    }

    @Test
    public void testListOfThrowsOnNullElement() {
        assertThrows(NullPointerException.class, () -> Java8Compat.listOf("a", null));
    }

    @Test
    public void testSetOfThrowsOnNullElement() {
        assertThrows(NullPointerException.class, () -> Java8Compat.setOf("a", null));
    }

    @Test
    public void testMapOfNoArgs() {
        Map<String, Integer> map = Java8Compat.mapOf();
        assertTrue(map.isEmpty());
        assertThrows(UnsupportedOperationException.class, () -> map.put("a", 1));
    }

    @Test
    public void testMapOfLargerArities() {
        Map<String, Integer> three = Java8Compat.mapOf("a", 1, "b", 2, "c", 3);
        assertEquals(3, three.size());
        assertEquals(3, three.get("c"));

        Map<String, Integer> four = Java8Compat.mapOf("a", 1, "b", 2, "c", 3, "d", 4);
        assertEquals(4, four.size());
        assertEquals(4, four.get("d"));

        Map<String, Integer> five = Java8Compat.mapOf("a", 1, "b", 2, "c", 3, "d", 4, "e", 5);
        assertEquals(5, five.size());
        assertEquals(5, five.get("e"));

        assertThrows(IllegalArgumentException.class,
                () -> Java8Compat.mapOf("a", 1, "b", 2, "a", 3));
        assertThrows(UnsupportedOperationException.class, () -> five.put("f", 6));
    }

    @Test
    public void testMapEntry() {
        Map.Entry<String, Integer> entry = Java8Compat.mapEntry("a", 1);
        assertEquals("a", entry.getKey());
        assertEquals(1, entry.getValue());
        assertThrows(UnsupportedOperationException.class, () -> entry.setValue(2));
        assertThrows(NullPointerException.class, () -> Java8Compat.mapEntry(null, 1));
        assertThrows(NullPointerException.class, () -> Java8Compat.mapEntry("a", null));
    }

    @Test
    public void testMapOfEntries() {
        Map<String, Integer> map = Java8Compat.mapOfEntries(
                Java8Compat.mapEntry("a", 1), Java8Compat.mapEntry("b", 2));
        assertEquals(1, map.get("a"));
        assertEquals(2, map.get("b"));
        assertThrows(UnsupportedOperationException.class, () -> map.put("c", 3));
        assertThrows(IllegalArgumentException.class, () -> Java8Compat.mapOfEntries(
                Java8Compat.mapEntry("a", 1), Java8Compat.mapEntry("a", 2)));
    }

    @Test
    public void testSetCopyOfDeduplicatesSilently() {
        // Set.copyOf dedups without throwing, unlike Set.of
        Set<String> copy = Java8Compat.setCopyOf(Java8Compat.listOf("a", "a", "b"));
        assertEquals(2, copy.size());
    }

    @Test
    public void testStreamOfOptional() {
        assertEquals(Java8Compat.listOf("a"),
                Java8Compat.stream(Optional.of("a")).collect(Collectors.toList()));
        assertEquals(0, Java8Compat.stream(Optional.empty()).count());
    }

    @Test
    public void testStreamOfNullable() {
        assertEquals(Java8Compat.listOf("a"),
                Java8Compat.streamOfNullable("a").collect(Collectors.toList()));
        assertEquals(0, Java8Compat.streamOfNullable(null).count());
    }

    @Test
    public void testWriteStringThrowsOnUnpairedSurrogate() throws Exception {
        Path path = Files.createTempFile("java8compat", ".txt");
        try {
            // Files.writeString throws rather than substituting replacement chars
            assertThrows(CharacterCodingException.class,
                    () -> Java8Compat.writeString(path, "\ud800"));
        } finally {
            Files.deleteIfExists(path);
        }
    }

    @Test
    public void testSetCopyOfIsSnapshotNotView() {
        Set<String> src = new java.util.LinkedHashSet<>();
        src.add("a");
        Set<String> copy = Java8Compat.setCopyOf(src);
        src.add("b");
        assertEquals(1, copy.size());
        assertTrue(copy.contains("a"));
    }

    @Test
    public void testWriteStringForwardsAppendOption() throws Exception {
        // Both production call sites pass CREATE, APPEND; dropping the varargs
        // forwarding at the polyfill would silently break append semantics.
        Path path = Files.createTempFile("java8compat", ".txt");
        try {
            Java8Compat.writeString(path, "a\n", StandardOpenOption.CREATE, StandardOpenOption.APPEND);
            Java8Compat.writeString(path, "b\n", StandardOpenOption.CREATE, StandardOpenOption.APPEND);
            assertEquals("a\nb\n", new String(Files.readAllBytes(path), StandardCharsets.UTF_8));
        } finally {
            Files.deleteIfExists(path);
        }
    }

    @Test
    public void testOrDoesNotInvokeSupplierWhenPresent() {
        // JDK Optional.or is lazy: the supplier is never called when present.
        assertEquals("a", Java8Compat.or(Optional.of("a"), () -> {
            throw new AssertionError("supplier must not run when Optional is present");
        }).get());
    }

    @Test
    public void testGrabBag() {
        // Zero-arg factory overloads yield empty immutable collections.
        assertTrue(Java8Compat.listOf().isEmpty());
        assertTrue(Java8Compat.setOf().isEmpty());
        assertTrue(Java8Compat.mapOfEntries().isEmpty());

        // setCopyOf of an empty collection.
        assertTrue(Java8Compat.setCopyOf(new java.util.ArrayList<String>()).isEmpty());

        // isBlank on empty string.
        assertTrue(Java8Compat.isBlank(""));

        // setCopyOf rejects null element.
        Set<String> nullElem = new java.util.LinkedHashSet<>();
        nullElem.add(null);
        assertThrows(NullPointerException.class, () -> Java8Compat.setCopyOf(nullElem));

        // mapOf 1-arity positive path.
        Map<String, Integer> one = Java8Compat.mapOf("a", 1);
        assertEquals(1, one.size());
        assertEquals(1, one.get("a"));
        assertThrows(UnsupportedOperationException.class, () -> one.put("b", 2));
    }
}
