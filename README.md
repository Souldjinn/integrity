# integrity
(exploratory) integrity testing service

## Interactive Test
```
run.sh
```

## Personal Goals

 * Standardized and useful way of running integrity tests.

 * Open Source. Key aspect is getting as much simple functionality.


## Design Parameters, First Iteration

### Design Parameter: Targets.

Use Cases:

 * **Ad-hoc (One Arbitrary Target)** - For example, you want to diagnose an issue in
   a single resource.

    - The simplest way to approach this by accepting a target or list of targets in
      a post request. The problem is that such an approach doesn't benefit from any
      automation.

    - More complicated, but interoperable with the arbitrary target list model, would
      be creating a file and having the system grab it.

 * **Cohort (Fixed Target List)** - For example, a group of resources are designated
   as canaries (Canary Deployments) or as a statistical sample of a population- which
   could be used to measure the efficacy or regressions due to a fix, or improvements/
   journey over time in a cohort.

    - This lends itself well to a recurring task because the task itself never changes
      so the only consideration is how to run it over and over again.

    - Comparisons between test runs are doable because they'll always have the same
      data.

    - This approach is very easy to automate.

 * ** (Arbitrary Target List)** - For example, as resources are created they are checked
   for integrity and regressions.

    - Comparability of test runs is much more complicated as targets enter and exit
      the pool.

    - Acquiring a target list becomes an issue and isn't simple to automate. The system
      can either:

      a. Make a pull to get the list from somewhere else; this is problematic because
         it pushes the logic about targets into another system (which shouldn't have to
         care), perhaps arbitrarily in a distributed systems case. Alternatively, it
         could be in S3.

      b. Have other systems push the target list over on a recurring basis. This is
         problematic for similar reasons.

Discussion:

 * My personal preference is to eliminate the arbitrary target list case while still
   figuring out how to support its use cases. In my judgement, we can get most of the
   benefits without the costs. Some ideas:

     - we could automate task creation by pulling new lists of resources regularly
       while cancelling old tasks.

     - we could create cohorts - lists of resources that we'd like to pass an integrity
       suite - then monitor their progress (for example, combined marketing and
       engineering fixes for key accounts) and apply high touch support. This is actually
       pretty good from a "having actionable goals" standpoint.

 * Getting rid of the ad-hoc case is more difficult, but is the same as a Cohort task
   that only has one resource and runs immediately.

     - we could provide a curl command that redirects and retries on task creation using
       a 202 Accepted and Redirected pattern:
        ```
        curl -XPOST --retry 3 --retry-delay 5 --retry-max-time 32 https://.../task
        ```

     - provided we've done a good job with cohorts and task setup, diagnostic information
       should become available for resources on a regular basis. The ad-hoc case becomes
       a matter of finding the right task and looking up the results.

Conclusion

 * Tasks (tests and targets) are created and can't be changed afterwards, but can be
   re-run.



### Design Parameter: Running Tasks.

Use cases:

 * **User Initiated** - someone wants to check the integrity of a resource. For example,
   someone on support diagnosing an issue. The best UX is entering a resource and a test
   suite and getting back a result synchronously.

 * **System Initiated** - a build system wants to make sure a deployment has gone well
   against canaries before full roll out. Options:

    - A synchronous call post-canary deploy

    - A callback hook to tell the system that everything is good. This is prone to
      network partitions unless we implement a message queue.

    - Drop results and let the system poll. For example, 202 Accepted with a Redirect to
      where the results will be- then the system polls until the resource has been
      created. This is prone to network partitions in a bad way if a response gets dropped.

    - Drop recent results in a given location with a timestamp. For example, the system
      sends off a request and polls a known location for status and timestamp. Once the
      location indicates fresh results are available, the system can pull them. This is
      not prone to nasty partition issues.

 * **Recurring** - a recurring task executes integrity tests to see if a herd of resources
   are doing okay. If they aren't, the system might report the results somewhere or
   initiate remediation steps (error correction and detection model). Options:

    - The 2nd and 4th options from System Initiated apply here.

    - Built in reporting and remediation. Personally, I think this overloads the scope
      of the design.

Discussion:

 * Stale data isn't all that interesting- recent results are actionable and we only need
   historical data to measure progress. This leads me to prefer letting clients poll a
   task page for fresh data (assuming timestamps are synced across services, a deploy
   tool would just check to see if the data is clean and has a time later than the deploy
   time).

 * Recurring runs that aren't resource intensive might run frequently enough to eliminate
   the need for system initiated calls. A continuous deploy pipeline that pushes out code
   every 30sec might be a problem, but releasing every hour or so is doable.

Conclusion:

 * Scheduled (or recurring) task runs ought to cover enough use-cases, provided they
   aren't too resource intensive.

 * Recent results for a task ought to be reported with the task, so other systems can
   just poll the task at a known location.

    - Interestingly, this means results can be cached! Of course, we may need to retain
      detailed historical data (can't think of why, but you never know).

Intermission:

 * Tasks are a list of targets and tests and can't change after creation.

 * Tasks run on a fixed schedule.

 * The most recent results for a task can be polled.