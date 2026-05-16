package org.auth;

import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.springframework.boot.test.context.SpringBootTest;

@Disabled("Skipping context load test as it requires database/redis")
@SpringBootTest
class AuthApplicationTests {

	@Test
	void contextLoads() {
	}

}
