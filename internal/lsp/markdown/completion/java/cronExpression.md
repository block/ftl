Declare a cron job, autocompleting on a cron expression.

A cron job is an Empty verb that will be called on a schedule. Supports both cron expressions and duration format.

```java
@Cron("0 * * * *")
public class Hourly {
    public void run() throws Exception {
        // Run every hour
    }
}

@Cron("6h")
public class EverySixHours {
    public void run() throws Exception {
        // Run every 6 hours
    }
}

@Cron("Mon")
public class Mondays {
    public void run() throws Exception {
        // Run every Monday
    }
}
```

See https://block.github.io/ftl/docs/reference/cron/
---

@Cron("${2:Minutes} ${3:Hours} ${4:DayOfMonth} ${5:Month} ${6:DayOfWeek}")
public class ${1:Name} {
    public void run() throws Exception {
        ${7:// TODO: Implement}
    }
} 
