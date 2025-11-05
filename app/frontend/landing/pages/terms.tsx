import React, { useEffect, useState } from "react";

type CurrencyResp = { currency?: string } | null;

export default function TermsPage() {
  const [currency, setCurrency] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;

    async function getCurrency() {
      try {
        const res = await fetch("https://api.bareuptime.co/ip/currency", {
          headers: { "Content-Type": "application/json" },
        });
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const json: CurrencyResp = await res.json();
        if (!mounted) return;
        setCurrency(json?.currency ?? null);
      } catch (err: any) {
        console.error(err);
        if (mounted) setError(String(err?.message ?? err));
      } finally {
        if (mounted) setLoading(false);
      }
    }

    getCurrency();
    return () => {
      mounted = false;
    };
  }, []);

  function renderOutsideTerms() {
    const highlightBox: React.CSSProperties = {
      background: '#fff7f0',
      border: '1px solid #ffd1a9',
      padding: 16,
      borderRadius: 6,
      marginTop: 8,
      marginBottom: 16,
    };

    const h1Style: React.CSSProperties = { marginTop: 0 };

    return (
      <div>
        <h1>Terms of Service</h1>
        <p>Effective Date: June 24, 2025</p>
        <p>Last Updated: June 24, 2025</p>

        <h2>1. Agreement to Terms</h2>
        <p>
          By accessing or using BareUptime ("Service"), you agree to be bound by
          these Terms of Service ("Terms"). If you disagree with any part of
          these terms, you may not access the Service.
        </p>

        <h2>2. Company Information</h2>
        <p>
          Service Provider: Penify Technologies LLC<br />
          Registered Address: 30 N Gould St Ste N, Sheridan, WY 82801<br />
          Email: support@hith.chat<br />
          Website: https://hith.chat
        </p>

        <h2>3. Description of Service</h2>
        <p>
          BareUptime is a website monitoring service that monitors website
          availability and performance, sends alerts when websites go down or
          experience issues, provides uptime statistics and reporting, offers
          SSL certificate monitoring, and delivers notifications through
          various channels (email, mobile apps, webhooks, etc.).
        </p>

        <h2>4. Account Registration</h2>
        <h3>4.1 Account Creation</h3>
        <ul>
          <li>You must provide accurate and complete information when creating an account.</li>
          <li>You are responsible for maintaining the confidentiality of your account credentials.</li>
          <li>You must be at least 13 years old to use the Service.</li>
          <li>One account per person or organization.</li>
        </ul>

        <h3>4.2 Account Security</h3>
        <p>
          You are responsible for all activities that occur under your account.
          Notify us immediately of any unauthorized use of your account. We
          reserve the right to suspend accounts that appear to be compromised.
        </p>

        <h2>5. Acceptable Use Policy</h2>
        <h3>5.1 Permitted Uses</h3>
        <p>Monitor websites that you own or have permission to monitor and use the Service for legitimate business purposes while complying with applicable laws.</p>

        <h3>5.2 Prohibited Uses</h3>
        <p>You may not:</p>
        <ul>
          <li>Monitor websites without proper authorization.</li>
          <li>Use the Service to harass, abuse, or harm others.</li>
          <li>Attempt to gain unauthorized access to our systems.</li>
          <li>Violate any laws or regulations.</li>
          <li>Interfere with or disrupt the Service.</li>
          <li>Use the Service for illegal activities or to monitor illegal content.</li>
          <li>Resell or redistribute the Service without permission.</li>
          <li>Reverse engineer or attempt to extract source code.</li>
          <li>Create multiple accounts to circumvent service limits.</li>
        </ul>

        <h2>6. Subscription Plans and Pricing</h2>
        <h3>6.1 Service Plans</h3>
        <p>
          Free Plan: Up to 50 monitors with basic features. Paid Plans: Starting at $15/year with additional features and monitors. Plan details and current pricing available at https://hith.chat
        </p>

        <h3>6.2 Payment Terms</h3>
        <p>
          Annual subscriptions are billed in advance. All fees are non-refundable
          except as required by law. Prices may change with 30 days' notice to
          existing subscribers. Failed payments may result in service
          suspension.
        </p>

        <h2>7. Service Availability and Performance</h2>
        <h3>7.1 Service Level</h3>
        <p>
          We strive to provide reliable monitoring with high uptime. Service
          availability is not guaranteed and may be affected by maintenance,
          outages, or technical issues. We do not guarantee 100% accuracy of
          monitoring results.
        </p>

        <h3>7.2 Monitoring Frequency</h3>
        <p>
          Free plans: Monitoring every 5 minutes. Paid plans: Monitoring every
          1-3 minutes (depending on plan). We reserve the right to adjust
          monitoring frequency based on service load.
        </p>

        <h2>8. Disclaimers and Limitation of Liability</h2>
        <h3>8.1 Service Disclaimers</h3>
        <p>
          THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES
          OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
          WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND
          NON-INFRINGEMENT.
        </p>

        <h3 style={{ color: '#7a1f1f' }}>8.2 Limitation of Liability</h3>
        <div style={highlightBox}>
          <p style={{ margin: 0, fontWeight: 600 }}>
            TO THE MAXIMUM EXTENT PERMITTED BY LAW, PENIFY TECHNOLOGIES LLC SHALL
            NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL,
            OR PUNITIVE DAMAGES, INCLUDING BUT NOT LIMITED TO LOSS OF PROFITS,
            DATA, OR BUSINESS INTERRUPTION.
          </p>
        </div>

        <h2>9. Termination</h2>
        <h3>9.1 Termination by You</h3>
        <p>
          You may terminate your account at any time through your account
          settings. Termination does not relieve you of payment obligations for
          services already provided.
        </p>

        <h3>9.2 Termination by Us</h3>
        <p>
          We may terminate or suspend your account immediately if you violate
          these Terms, engage in fraudulent activity, fail to pay required
          fees, or use the Service in a way that could harm us or other users.
        </p>

        <h2>10. Governing Law and Disputes</h2>
        <p>
          These Terms are governed by the laws of the State of Wyoming,
          United States, without regard to conflict of law principles. Any
          disputes will be resolved through binding arbitration in Wyoming. You
          waive the right to participate in class action lawsuits. Wyoming
          state courts have exclusive jurisdiction for any matters not subject
          to arbitration.
        </p>

        <h2>11. Contact Information</h2>
        <p>
          For questions about these Terms of Service, contact us:<br />
          Email: legal@hith.chat<br />
          Support: support@hith.chat<br />
          Address: Penify Technologies LLC, 30 N Gould St Ste N, Sheridan, WY 82801
        </p>

        <p>By using BareUptime, you acknowledge that you have read, understood, and agree to be bound by these Terms of Service.</p>
      </div>
    );
  }

  function renderIndiaTerms() {
    const highlightBox: React.CSSProperties = {
      background: '#fff7f0',
      border: '1px solid #ffd1a9',
      padding: 16,
      borderRadius: 6,
      marginTop: 8,
      marginBottom: 16,
    };

    return (
      <div>
        <h1>Terms of Service</h1>
        <p>Last Updated: February 05, 2024</p>

        <p>
          These terms and conditions ("Agreement") set forth the general terms
          and conditions of your use of the penify.dev website ("Website" or
          "Service") and any of its related products and services (collectively,
          "Services"). This Agreement is legally binding between you
          ("User", "you" or "your") and Snorkell Associates and Co ("Snorkell
          Associates and Co", "we", "us" or "our"). If you are entering into
          this Agreement on behalf of a business or other legal entity, you
          represent that you have the authority to bind such entity to this
          Agreement, in which case the terms "User", "you" or "your" shall
          refer to such entity. If you do not have such authority, or if you do
          not agree with the terms of this Agreement, you must not accept this
          Agreement and may not access and use the Website and Services. By
          accessing and using the Website and Services, you acknowledge that
          you have read, understood, and agree to be bound by the terms of this
          Agreement. You acknowledge that this Agreement is a contract between
          you and Snorkell Associates and Co, even though it is electronic and
          is not physically signed by you, and it governs your use of the
          Website and Services.
        </p>

        <h2>Table of Contents</h2>
        <ul>
          <li>Accounts and membership</li>
          <li>User content</li>
          <li>Backups</li>
          <li>Links to other resources</li>
          <li>Prohibited uses</li>
          <li>Intellectual property rights</li>
          <li>Limitation of liability</li>
          <li>Fees and Pricing</li>
          <li>Indemnification</li>
          <li>Severability</li>
          <li>Dispute resolution</li>
          <li>Changes and amendments</li>
          <li>Acceptance of these terms</li>
          <li>Contacting us</li>
        </ul>

        <h3>Accounts and Membership</h3>
        <p>
          You must be at least 13 years of age to use the Website and Services.
          By using the Website and Services and by agreeing to this Agreement
          you warrant and represent that you are at least 13 years of age. If
          you create an account on the Website, you are responsible for
          maintaining the security of your account and you are fully
          responsible for all activities that occur under the account and any
          other actions taken in connection with it. We may, but have no
          obligation to, monitor and review new accounts before you may sign in
          and start using the Services. Providing false contact information of
          any kind may result in the termination of your account. You must
          immediately notify us of any unauthorized uses of your account or any
          other breaches of security. We will not be liable for any acts or
          omissions by you, including any damages of any kind incurred as a
          result of such acts or omissions. We may suspend, disable, or delete
          your account (or any part thereof) if we determine that you have
          violated any provision of this Agreement or that your conduct or
          content would tend to damage our reputation and goodwill. If we
          delete your account for the foregoing reasons, you may not re-register
          for our Services. We may block your email address and Internet
          protocol address to prevent further registration.
        </p>

        <h3>User Content</h3>
        <p>
          We do not own any data, information or material (collectively,
          "Content") that you submit on the Website in the course of using the
          Service. You shall have sole responsibility for the accuracy,
          quality, integrity, legality, reliability, appropriateness, and
          intellectual property ownership or right to use of all submitted
          Content. We may monitor and review the Content on the Website
          submitted or created using our Services by you. You grant us
          permission to access, copy, distribute, store, transmit, reformat,
          display and perform the Content of your user account solely as
          required for the purpose of providing the Services to you. Without
          limiting any of those representations or warranties, we have the
          right, though not the obligation, to, in our own sole discretion,
          refuse or remove any Content that, in our reasonable opinion, violates
          any of our policies or is in any way harmful or objectionable. You
          also grant us the license to use, reproduce, adapt, modify, publish
          or distribute the Content created by you or stored in your user
          account for commercial, marketing or any similar purpose.
        </p>

        <h3>Backups</h3>
        <p>
          We are not responsible for the Content residing on the Website. In no
          event shall we be held liable for any loss of any Content. It is your
          sole responsibility to maintain appropriate backup of your Content.
          Notwithstanding the foregoing, on some occasions and in certain
          circumstances, with absolutely no obligation, we may be able to
          restore some or all of your data that has been deleted as of a
          certain date and time when we may have backed up data for our own
          purposes. We make no guarantee that the data you need will be
          available.
        </p>

        <h3>Links to Other Resources</h3>
        <p>
          Although the Website and Services may link to other resources (such
          as websites, mobile applications, etc.), we are not, directly or
          indirectly, implying any approval, association, sponsorship,
          endorsement, or affiliation with any linked resource, unless
          specifically stated herein. We are not responsible for examining or
          evaluating, and we do not warrant the offerings of, any businesses or
          individuals or the content of their resources. We do not assume any
          responsibility or liability for the actions, products, services, and
          content of any other third parties. You should carefully review the
          legal statements and other conditions of use of any resource which
          you access through a link on the Website. Your linking to any other
          off-site resources is at your own risk.
        </p>

        <h3>Prohibited Uses</h3>
        <p>
          In addition to other terms as set forth in the Agreement, you are
          prohibited from using the Website and Services or Content: (a) for
          any unlawful purpose; (b) to solicit others to perform or participate
          in any unlawful acts; (c) to violate any international, federal,
          provincial or state regulations, rules, laws, or local ordinances; (d)
          to infringe upon or violate our intellectual property rights or the
          intellectual property rights of others; (e) to harass, abuse, insult,
          harm, defame, slander, disparage, intimidate, or discriminate based
          on gender, sexual orientation, religion, ethnicity, race, age,
          national origin, or disability; (f) to submit false or misleading
          information; (g) to upload or transmit viruses or any other type of
          malicious code that will or may be used in any way that will affect
          the functionality or operation of the Website and Services, third
          party products and services, or the Internet; (h) to spam, phish,
          pharm, pretext, spider, crawl, or scrape; (i) for any obscene or
          immoral purpose; or (j) to interfere with or circumvent the security
          features of the Website and Services, third party products and
          services, or the Internet. We reserve the right to terminate your use
          of the Website and Services for violating any of the prohibited uses.
        </p>

        <h3>Intellectual Property Rights</h3>
        <p>
          "Intellectual Property Rights" means all present and future rights
          conferred by statute, common law or equity in or in relation to any
          copyright and related rights, trademarks, designs, patents,
          inventions, goodwill and the right to sue for passing off, rights to
          inventions, rights to use, and all other intellectual property rights,
          in each case whether registered or unregistered and including all
          applications and rights to apply for and be granted, rights to claim
          priority from, such rights and all similar or equivalent rights or
          forms of protection and any other results of intellectual activity
          which subsist or will subsist now or in the future in any part of the
          world. This Agreement does not transfer to you any intellectual
          property owned by Snorkell Associates and Co or third parties, and
          all rights, titles, and interests in and to such property will remain
          (as between the parties) solely with Snorkell Associates and Co. All
          trademarks, service marks, graphics and logos used in connection with
          the Website and Services, are trademarks or registered trademarks of
          Snorkell Associates and Co or its licensors. Other trademarks, service
          marks, graphics and logos used in connection with the Website and
          Services may be the trademarks of other third parties. Your use of
          the Website and Services grants you no right or license to reproduce
          or otherwise use any of Snorkell Associates and Co or third party
          trademarks.
        </p>

        <h3 style={{ color: '#7a1f1f' }}>Limitation of Liability</h3>
        <div style={highlightBox}>
          <p style={{ margin: 0, fontWeight: 600 }}>
            To the fullest extent permitted by applicable law, in no event will
            Snorkell Associates and Co, its affiliates, directors, officers,
            employees, agents, suppliers or licensors be liable to any person for
            any indirect, incidental, special, punitive, cover or consequential
            damages (including, without limitation, damages for lost profits,
            revenue, sales, goodwill, use of content, impact on business, business
            interruption, loss of anticipated savings, loss of business
            opportunity) however caused, under any theory of liability, including,
            without limitation, contract, tort, warranty, breach of statutory
            duty, negligence or otherwise, even if the liable party has been
            advised as to the possibility of such damages or could have foreseen
            such damages. To the maximum extent permitted by applicable law, the
            aggregate liability of Snorkell Associates and Co and its affiliates,
            officers, employees, agents, suppliers and licensors relating to the
            services will be limited to an amount no greater than one dollar or
            any amounts actually paid in cash by you to Snorkell Associates and
            Co for the prior one month period prior to the first event or
            occurrence giving rise to such liability. The limitations and
            exclusions also apply if this remedy does not fully compensate you for
            any losses or fails of its essential purpose.
          </p>
        </div>

        <h3>Fees and Pricing</h3>
        <p>
          The fees and pricing for the services availed from the Company shall
          be specified in the subscription plans ("Subscription") available on
          the Platform or in the work order form executed by the Parties. You
          will be billed in advance on a recurring and periodic basis
          ("Billing Cycle") based on the subscription plan selected by you. At
          the end of each Billing Cycle, your Subscription will automatically
          renew under the exact same conditions unless you cancel it or Company
          cancels it. You may cancel your Subscription renewal on the Platform.
          A valid payment method, like credit card, is required to process the
          payment for your subscription. You shall provide Company with
          accurate and complete billing information including full name,
          address, state, zip code, telephone number, and a valid payment
          method information. By submitting such payment information, you
          automatically authorise the Company to charge all Subscription fees
          incurred through your account to any such payment instruments. Should
          automatic billing fail to occur for a reason, Company will issue an
          electronic invoice indicating that you must proceed manually, within
          a certain deadline date, with the full payment corresponding to the
          billing period as indicated on the invoice.
        </p>

        <h3>Indemnification</h3>
        <p>
          You agree to indemnify and hold Snorkell Associates and Co and its
          affiliates, directors, officers, employees, agents, suppliers and
          licensors harmless from and against any liabilities, losses, damages
          or costs, including reasonable attorneys' fees, incurred in
          connection with or arising from any third party allegations, claims,
          actions, disputes, or demands asserted against any of them as a
          result of or relating to your Content, your use of the Website and
          Services or any willful misconduct on your part.
        </p>

        <h3>Severability</h3>
        <p>
          All rights and restrictions contained in this Agreement may be
          exercised and shall be applicable and binding only to the extent that
          they do not violate any applicable laws and are intended to be
          limited to the extent necessary so that they will not render this
          Agreement illegal, invalid or unenforceable. If any provision or
          portion of any provision of this Agreement shall be held to be
          illegal, invalid or unenforceable by a court of competent
          jurisdiction, it is the intention of the parties that the remaining
          provisions or portions thereof shall constitute their agreement with
          respect to the subject matter hereof, and all such remaining
          provisions or portions thereof shall remain in full force and
          effect.
        </p>

        <h3>Dispute Resolution</h3>
        <p>
          The formation, interpretation, and performance of this Agreement and
          any disputes arising out of it shall be governed by the substantive
          and procedural laws of Rajasthan, India without regard to its rules on
          conflicts or choice of law and, to the extent applicable, the laws of
          India. The exclusive jurisdiction and venue for actions related to the
          subject matter hereof shall be the courts located in Rajasthan, India,
          and you hereby submit to the personal jurisdiction of such courts. You
          hereby waive any right to a jury trial in any proceeding arising out
          of or related to this Agreement. The United Nations Convention on
          Contracts for the International Sale of Goods does not apply to this
          Agreement.
        </p>

        <h3>Changes and Amendments</h3>
        <p>
          We reserve the right to modify this Agreement or its terms related to
          the Website and Services at any time at our discretion. When we do,
          we will revise the updated date at the bottom of this page, post a
          notification on the main page of the Website, send you an email to
          notify you. We may also provide notice to you in other ways at our
          discretion, such as through the contact information you have provided.
        </p>

        <h3>Acceptance of These Terms</h3>
        <p>
          You acknowledge that you have read this Agreement and agree to all
          its terms and conditions. By accessing and using the Website and
          Services you agree to be bound by this Agreement. If you do not agree
          to abide by the terms of this Agreement, you are not authorized to
          access or use the Website and Services. This terms and conditions
          policy was created with the help of WebsitePolicies.
        </p>

        <h3>Contacting Us</h3>
        <p>
          If you have any questions, concerns, or complaints regarding this
          Agreement, we encourage you to contact us using the details below:

          <br /> Email: legal@hith.chat<br /> Support: support@hith.chat<br /> Address: Snorkell Associates and Co, Rajasthan, India
        </p>

        <p>This document was last updated on February 05, 2024</p>
      </div>
    );
  }

  return (
    <div style={{ padding: 20, display: 'flex', justifyContent: 'center' }}>
      <div style={{ maxWidth: 900, width: '100%' }}>
        {loading && <p>Loading terms of service...</p>}
        {error && <p style={{ color: "red" }}>Failed to load terms: {error}</p>}
        {!loading && !error && (
          <div>
            {currency === "INR" ? renderIndiaTerms() : renderOutsideTerms()}
          </div>
        )}
      </div>
    </div>
  );
}
