package org.openapis.clientcredentials.openapi.hooks;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.time.OffsetDateTime;
import java.time.temporal.ChronoUnit;
import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.Test;
import org.openapis.clientcredentials.openapi.utils.SessionManager;
import org.openapis.clientcredentials.openapi.utils.SessionManager.HasSessionKey;
import org.openapis.clientcredentials.openapi.utils.SessionManager.Session;

public class ClientCredentialsHookAdditionalTest {

    @Test
    public void testNotExpiredWhenNoTokenExpiryKnown() {
        // If the response did not include a `expires_in` value, then the token is considered to never expire.
        assertFalse(SessionManager.hasTokenExpired(Optional.empty(), OffsetDateTime.now()));
    }

    @Test
    public void testNotExpiredWhenTokenExpiryMoreThanThresholdSecondsInFuture() {
        OffsetDateTime now = OffsetDateTime.now();
        OffsetDateTime expiry = now.plus(SessionManager.REFRESH_BEFORE_EXPIRY_SECONDS + 1, ChronoUnit.SECONDS);
        assertFalse(SessionManager.hasTokenExpired(Optional.of(expiry), now));
    }

    @Test
    public void testIsExpiredWhenTokenExpiryLessThanThresholdSecondsInFuture() {
        OffsetDateTime now = OffsetDateTime.now();
        OffsetDateTime expiry = now.plus(SessionManager.REFRESH_BEFORE_EXPIRY_SECONDS / 2, ChronoUnit.SECONDS);
        assertTrue(SessionManager.hasTokenExpired(Optional.of(expiry), now));
    }

    @Test
    public void testShouldCreateNewSessionIfSessionPresentAndScopesNotMatched() {
        Session<MyCredentials> session = new Session<>(new MyCredentials(), //
                Optional.of("token"), //
                List.of("read", "write"), //
                Optional.of(OffsetDateTime.now().plus(1, ChronoUnit.DAYS)));

        // Should create new session when scopes don't match
        assertFalse(SessionManager.hasRequiredScopes(session.scopes(), List.of("read", "explode")));
        assertFalse(SessionManager.hasTokenExpired(session.expiresAt(), OffsetDateTime.now()));

        // Should not create new session when scopes match
        assertTrue(SessionManager.hasRequiredScopes(session.scopes(), List.of("read")));
        assertFalse(SessionManager.hasTokenExpired(session.expiresAt(), OffsetDateTime.now()));
    }

    @Test
    public void testShouldCreateNewSessionIfSessionPresentAndAllScopesFoundInSession() {
        Session<MyCredentials> session = new Session<>(new MyCredentials(), //
                Optional.of("token"), //
                List.of("read", "write"), //
                Optional.of(OffsetDateTime.now().plus(1, ChronoUnit.DAYS)));

        // Should not create new session when all required scopes are present and not expired
        assertTrue(SessionManager.hasRequiredScopes(session.scopes(), List.of("read")));
        assertFalse(SessionManager.hasTokenExpired(session.expiresAt(), OffsetDateTime.now()));
    }

    @Test
    public void testShouldCreateNewSessionIfSessionPresentButExpired() {
        Session<MyCredentials> session = new Session<>(new MyCredentials(), //
                Optional.of("token"), //
                List.of("read", "write"), //
                Optional.of(OffsetDateTime.now().plus(-1, ChronoUnit.DAYS)));

        // Should create new session when token is expired, even if scopes match
        assertTrue(SessionManager.hasRequiredScopes(session.scopes(), List.of("read")));
        assertTrue(SessionManager.hasTokenExpired(session.expiresAt(), OffsetDateTime.now()));
    }

    public static final class MyCredentials implements HasSessionKey {

        @Override
        public String sessionKey() {
            return "key";
        }

    }
}
