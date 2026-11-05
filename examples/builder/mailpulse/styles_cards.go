package main

// cardsFixCSS separates stacked sections into individual cards on the
// Overview, Campaigns, Content, engagement, revenue, and analytics tabs:
// 16px gaps between sections, 12px inner radii, and 20px 24px inner padding,
// with the old hairline dividers removed where cards separate. Grids inside
// .mp-stack and .mp-ov hold one .of-card per cell, so every cell keeps its
// full border and 12px curves, including the second cell. The integrator
// concatenates this sheet into the theme after appCSS.
const cardsFixCSS = `
.mp-card.mp-ov .of-card,
.mp-card.mp-camp .of-card,
.mp-card.mp-cont .of-card,
.mp-card.mp-stack .of-card {
  border-radius: 12px;
  padding: 20px 24px;
  box-shadow: none;
}
.mp-card.mp-ov,
.mp-card.mp-camp,
.mp-card.mp-cont {
  overflow: hidden;
}
.mp-card.mp-stack {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
  overflow: hidden;
}
.mp-card.mp-stack > .mp-pagehead {
  background: #fff;
  border: 1px solid #e8e8e6;
  border-radius: 12px;
  padding: 20px 24px;
}
.mp-card.mp-stack > .mp-kpis {
  background: #fff;
  border: 1px solid #e8e8e6;
  border-radius: 12px;
  overflow: hidden;
}
.mp-card.mp-stack > .mp-grid-mid,
.mp-card.mp-stack > .mp-grid-bot {
  gap: 16px;
  border: 0;
  background: transparent;
}
.mp-card.mp-stack > .mp-grid-mid > *,
.mp-card.mp-stack > .mp-grid-bot > * {
  background: #fff;
  border: 1px solid #e8e8e6;
  border-radius: 12px;
  padding: 20px 24px;
  box-shadow: none;
}
.mp-card.mp-ov {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
}
.mp-ov > .mp-pagehead {
  background: #fff;
  border: 1px solid #e8e8e6;
  border-radius: 12px;
  padding: 20px 24px;
}
.mp-ov > .mp-kpis {
  background: #fff;
  border: 1px solid #e8e8e6;
  border-radius: 12px;
  overflow: hidden;
}
.mp-ov .mp-kpi {
  padding: 20px;
}
.mp-ov > .mp-grid-mid,
.mp-ov > .mp-grid-bot {
  gap: 16px;
  border: 0;
  background: transparent;
}
.mp-ov > .mp-grid-mid > *,
.mp-ov > .mp-grid-bot > * {
  background: #fff;
  border: 1px solid #e8e8e6;
  border-radius: 12px;
  padding: 20px 24px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}
.mp-ov > .mp-grid-mid > *:first-child {
  display: flex;
  flex-direction: column;
}
.mp-ov > .mp-grid-mid > *:first-child .of-card-content {
  display: flex;
  flex-direction: column;
  flex: 1;
}
.mp-ov > .mp-grid-mid > *:first-child .mp-bars {
  flex: 1;
}
.mp-ov > .mp-grid-mid > *:nth-child(2),
.mp-ov > .mp-grid-bot > *:nth-child(2) {
  border-left: 1px solid #e8e8e6;
}
.mp-camp > .mp-heat-badges {
  padding: 0 24px;
  margin: 0 0 16px;
}
.mp-camp > .of-card {
  margin: 0 24px 24px;
}
.mp-cont > .mp-grid-bot {
  gap: 16px;
  padding: 0 24px 24px;
  border: 0;
  background: transparent;
}
.mp-cont > .mp-grid-bot > * {
  background: #fff;
  border: 1px solid #e8e8e6;
  border-radius: 12px;
  padding: 20px 24px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}
.mp-cont > .mp-grid-bot > *:nth-child(2) {
  border-left: 1px solid #e8e8e6;
}
`
