# Concepts

## Actor & Action

An `Actor` is any entity that performs some actions. In a real-world use-case, this can represent end-users of some
product.

```json
{
  "id": "user:1234",
  "attribs": {
    "active": true,
    "email": "user1234@themail.com"
  }
}
```

An `Action` is an event describing an action performed by an `Actor`. In the above actor example, actions can be
purchasing an item, transacting or using soem feature of the product, etc.

```json
{
  "id": "register-user1234",
  "time": "2022-02-03T10:59:25.000000",
  "actor_id": "user:1234",
  "data": {
    "type": "REGISTER_ACCOUNT"
  }
}
```

## Campaign

A campaign represents set of steps that an `Actor` may complete by doing actions.

A sample `Campaign`:

```json
{
  "name": "a-sample-campaign",
  "scope": [
    "team:business",
    "country:us"
  ],
  "enabled": true,
  "start_at": "2022-02-01T00:00:00Z",
  "end_at": "2022-02-11T23:59:59.9999Z",
  "priority": 0,
  "is_unordered": false,
  "steps": [
    "event.type == 'REGISTER_ACCOUNT'",
    "event.type == 'PURCHASE_ITEM' and event.amount >= 1000"
  ]
}
```

Few key points from this:

* This campaign remains active from 01-Feb-2022 (beginning of day) to 11-Feb-2022 (end of day).
* First step completes when user registers an account.
* Second step completes when user purchases an item with price amount of at-least 1000.

## Enrolment

An `Enrolment` is a binding between an actor and a campaign.

```json
{
  "actor_id": "user:123",
  "campaign_id": "a-sample-campaign",
  "status": "ACTIVE",
  "started_at": "2022-02-03T10:59:18.762481+05:30",
  "ends_at": "2022-02-11T10:51:40.432191+05:30",
  "remaining_steps": 1,
  "completed_steps": [
    {
      "step_id": 0,
      "done_at": "2022-02-03T10:59:25.000000",
      "action_id": "register-user:123"
    }
  ]
}
```

* Above enrolment represents, a binding between actor identified as `user:123` and campaign `a-sample-campaign`.
* It also shows the step that is already completed (i.e., step #0) and the remaining steps.