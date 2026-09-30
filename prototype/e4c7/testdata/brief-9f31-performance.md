# Brief: 9f31 Sensor board at performance

This brief stands alone. It covers one item of the work queue of ada@example.org on `main`: work waiting.

## The item

- Contributor: ada@example.org  (default)
- Reviewer: none
- Gate: ⚡ performance: Target performance shown in the deliverable's technology
- Junction references, which expand the criteria for this task:
  - Power budget, docs/sensor-board.md#power-budget
- Cause: blocked: Barometer ICs on 14-week backorder

## The task

The PCB that carries the barometer, hygrometer and thermometer and powers them from the solar cell within the power budget.

References:

- docs/sensor-board.md
- Barometer datasheet, https://example.org/datasheets/bmp390.pdf

Under: Sensor node (`4e2b`)

## Requirements

Requires: none.

Also needed by: `c07d` Node firmware from design to implementation, Pin map and sensor bus

## Status

function stalled (blocked), 2026-09-28, ada@example.org: Barometer ICs on 14-week backorder

## When done

Do not start. The work waits: blocked: Barometer ICs on 14-week backorder. Nothing is committed.

## Reproduce

```
tabloio queue --person ada@example.org --brief 9f31 performance
tabloio task 9f31
```
