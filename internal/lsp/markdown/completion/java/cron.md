Declare a cron job.

Cron jobs are scheduled functions that run on a recurring basis. The schedule can be specified using either cron syntax or duration syntax.

```java
import xyz.block.ftl.Cron;

class MyCron {
	// Run every 30 seconds
	@Cron("30s")
	void frequentJob() {
		// Frequent cron job logic
	}

	// Run at the start of every hour
	@Cron("0 * * * *")
	void hourly() {
		// Hourly cron job logic
	}
}
```

See https://block.github.io/ftl/docs/reference/cron/
---

class ${1:Name} {
	@Cron("${2:schedule}")
	void ${3:name}() {
		${4:// Add your cron job logic here}
	}
} 
