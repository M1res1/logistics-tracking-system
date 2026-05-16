package org.auth.util;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.test.util.ReflectionTestUtils;

import static org.junit.jupiter.api.Assertions.*;

public class SecurityFunctionsTest {

    private PasswordEncoder passwordEncoder;
    private JwtUtil jwtUtil;

    @BeforeEach
    void setUp() {
        passwordEncoder = new BCryptPasswordEncoder();
        jwtUtil = new JwtUtil();
        ReflectionTestUtils.setField(jwtUtil, "secret", "404E635266556A586E3272357538782F413F4428472B4B6250645367566B5970");
        ReflectionTestUtils.setField(jwtUtil, "accessExpiration", 3600000L);
        ReflectionTestUtils.setField(jwtUtil, "refreshExpiration", 86400000L);
    }

    @Test
    public void testPasswordHashing() {
        String password = "mySecretPassword";
        String hash = passwordEncoder.encode(password);
        
        assertNotNull(hash);
        assertNotEquals(password, hash);
        assertTrue(passwordEncoder.matches(password, hash));
        assertFalse(passwordEncoder.matches("wrongPassword", hash));
    }

    @Test
    public void testJwtSignAndVerify() {
        String email = "test@example.com";
        String userType = "ADMIN";
        Long userId = 1L;

        String token = jwtUtil.generateToken(email, userType, userId);
        assertNotNull(token);

        String extractedEmail = jwtUtil.extractUsername(token);
        assertEquals(email, extractedEmail);

        String extractedUserType = jwtUtil.extractClaim(token, claims -> claims.get("user_type", String.class));
        assertEquals(userType, extractedUserType);
        
        Long extractedUserId = jwtUtil.extractClaim(token, claims -> claims.get("user_id", Long.class));
        assertEquals(userId, extractedUserId);

        assertTrue(jwtUtil.isTokenValid(token, email));
        assertFalse(jwtUtil.isTokenExpired(token));
    }
}
