import { TemplateResult, html } from 'lit';
import { customElement, property } from 'lit/decorators.js';

import { BaseComponent } from '../../ts/base-component';
import violationDrawerStyles from './violation-drawer.styles';

@customElement('violation-drawer')
export class ViolationDrawerComponent extends BaseComponent {
  static styles = violationDrawerStyles;

  @property({ type: Element })
  code: Element;

  @property({ type: String })
  message: string;

  @property({ type: String })
  path: string;

  @property({ type: String })
  category: string;

  @property({ type: String })
  ruleId: string;

  @property({ type: String })
  howToFix: string;

  @property({ type: String })
  description: string;

  private _visible: boolean;

  private static replaceTicks(message: string): TemplateResult[] {
    const rx = /(`[^`]*`)/g;
    const sections = message.split(rx);
    const renders: Array<TemplateResult> = new Array<TemplateResult>();

    sections.forEach((section: string) => {
      if (section.match(rx)) {
        const bt = section.replace(/`/g, '');
        const dat = html`<span class="backtick-element">${bt}</span>`;
        renders.push(dat);
      } else {
        if (section != '') {
          renders.push(html`${section}`);
        }
      }
    });
    return renders;
  }

  render() {
    if (this._visible) {
      return html`
        <h2>${ViolationDrawerComponent.replaceTicks(this.message)}</h2>
        ${this.code}
        ${this.description ? html`
          <hr />
          <section class="description">
            <h3>${this.ruleId}</h3>
            <p>${this.description}</p>
          </section>
        ` : html`
          <hr />
          <h3>${this.ruleId}</h3>
        `}
        ${this.howToFix ? html`
          <hr />
          <section class="how-to-fix">
            <h3>How to fix violation</h3>
            <p>${this.howToFix}</p>
          </section>
        ` : ''}
      `;
    } else {
      return html`
        <section class="select-violation">
          <p>Please select a rule violation from a category.</p>
        </section>
      `;
    }
  }

  get drawer(): HTMLElement {
    return document.querySelector('violation-drawer') as HTMLElement;
  }

  public show() {
    this._visible = true;
    this.drawer.classList.add('drawer-active');
    this.requestUpdate();
  }

  public hide() {
    this._visible = false;
    this.drawer.classList.remove('drawer-active');
    this.requestUpdate();
  }
}
