package org.auth.util;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;

import java.util.Date;

import static org.junit.jupiter.api.Assertions.*;

class JwtUtilTest {

    private JwtUtil jwtUtil;
    private final String secret = "404E635266556A586E3272357538782F413F4428472B4B6250645367566B5970";
    private final long accessExpiration = 3600000;
    private final long refreshExpiration = 86400000;

    @BeforeEach
    void setUp() {
        jwtUtil = new JwtUtil();
        ReflectionTestUtils.setField(jwtUtil, "secret", secret);
        ReflectionTestUtils.setField(jwtUtil, "accessExpiration", accessExpiration);
        ReflectionTestUtils.setField(jwtUtil, "refreshExpiration", refreshExpiration);
    }

    @Test
    void generateToken_AndExtractUsername_Success() {
        String email = "test@example.com";
        String token = jwtUtil.generateToken(email, "CUSTOMER", 1L);

        assertNotNull(token);
        assertEquals(email, jwtUtil.extractUsername(token));
    }

    @Test
    void isTokenValid_Success() {
        String email = "test@example.com";
        String token = jwtUtil.generateToken(email, "CUSTOMER", 1L);

        assertTrue(jwtUtil.isTokenValid(token, email));
    }

    @Test
    void isTokenValid_InvalidEmail_ReturnsFalse() {
        String email = "test@example.com";
        String token = jwtUtil.generateToken(email, "CUSTOMER", 1L);

        assertFalse(jwtUtil.isTokenValid(token, "wrong@example.com"));
    }

    @Test
    void generateRefreshToken_Success() {
        String email = "test@example.com";
        String token = jwtUtil.generateRefreshToken(email);

        assertNotNull(token);
        assertEquals(email, jwtUtil.extractUsername(token));
    }

    @Test
    void extractExpiration_Success() {
        String email = "test@example.com";
        String token = jwtUtil.generateToken(email, "CUSTOMER", 1L);

        Date expiration = jwtUtil.extractExpiration(token);
        assertTrue(expiration.after(new Date()));
    }

    @Test
    void isTokenExpired_ReturnsFalseForNewToken() {
        String email = "test@example.com";
        String token = jwtUtil.generateToken(email, "CUSTOMER", 1L);

        assertFalse(jwtUtil.isTokenExpired(token));
    }
}
