import React from 'react';
import Router from 'next/router';
import cookies from 'next-cookies';
import { Segment, Header, Button, Form, Message } from 'semantic-ui-react';
import Client from '../lib/client';
import Nav from '../components/nav';
import Footer from '../components/footer';
import PageHeader from '../components/pageheader';

export default class Copyright extends React.Component {
  async getInitialProps(ctx) {
    const { authTokenID } = cookies(ctx);
    return { authTokenID };
  }

  render() {
    return (
      <div>
        <PageHeader />
        <div className="holder">
          <Nav activeItem="copyright" authAccount={null} />
          <Segment padded>
            <Header as="h1">SiaCDN - Copyright Policy</Header>
            <p>Last Updated: October 7, 2017</p>
            <p>
              SiaCDN respects the intellectual property rights of others and
              expects its users to do the same.
            </p>
            <p>
              In accordance with the Digital Millennium Copyright Act of 1998
              (&ldquo;DMCA&rdquo;), the text of which may be found on the U.S.
              Copyright Office website at{' '}
              <a href="http://www.copyright.gov/legislation/pl105-304.pdf">
                http://www.copyright.gov/legislation/pl105-304.pdf
              </a>, and other applicable laws, SiaCDN has adopted a policy of
              terminating, in appropriate circumstances and at SiaCDN&rsquo;s
              sole discretion, the accounts of users who are deemed to be repeat
              infringers. SiaCDN may also, at its sole discretion, limit access
              to SiaCDN&rsquo;s website and services (collectively,
              &ldquo;Service&rdquo;) and/or terminate the accounts of any users
              who infringe any intellectual property rights of others, whether
              or not there is any repeat infringement. SiaCDN will respond to
              claims of copyright infringement committed on the SiaCDN website
              that are reported to SiaCDN&rsquo;s Designated Copyright Agent,
              identified in the sample notice below.
            </p>
            <p>
              If you knowingly misrepresent in your notification that the
              material or activity is infringing, you will be liable for any
              damages, including costs and attorneys&rsquo; fees, incurred by us
              or the alleged infringer as the result of our relying upon such
              misrepresentation in removing or disabling access to the material
              or activity claimed to be infringing.
            </p>
            <p>
              If you are a copyright owner, or are authorized to act on behalf
              of one, or authorized to act under any exclusive right under
              copyright, please report alleged copyright infringements taking
              place on or through the SiaCDN Site by completing the following
              DMCA Notice of Alleged Infringement and delivering it to
              SiaCDN&rsquo;s Designated Copyright Agent. Upon receipt of the
              Notice as described below, SiaCDN will take whatever action, in
              its sole discretion, it deems appropriate, including removal of
              the challenged material from the SiaCDN Site.
            </p>

            <Header as="h2">
              DMCA Notice of Alleged Infringement (&ldquo;Notice&rdquo;):
            </Header>
            <ol>
              <li>
                Identify the copyrighted work that you claim has been infringed,
                or if multiple copyrighted works are covered by this Notice you
                may provide a representative list of the copyrighted works that
                you claim have been infringed.
              </li>
              <li>
                Identify the material that you claim is infringing (or to be the
                subject of infringing activity) and that is to be removed or
                access to which is to be disabled, and information reasonably
                sufficient to permit us to locate the material, including at a
                minimum, if applicable, the URL of the link shown on the SiaCDN
                Site where such material may be found.
              </li>
              <li>
                Provide your mailing address, telephone number, and, if
                available, email address.
              </li>
              <li>
                Include both of the following statements in the body of the
                Notice: (i) &ldquo;I hereby state that I have a good faith
                belief that the disputed use of the copyrighted material is not
                authorized by the copyright owner, its agent, or the law (e.g.,
                as a fair use)&rdquo;; and (ii) &ldquo;I hereby state that the
                information in this Notice is accurate and, under penalty of
                perjury, that I am the owner, or authorized to act on behalf of
                the owner, of the copyright or of an exclusive right under the
                copyright that is allegedly infringed.&rdquo;
              </li>
              <li>
                Provide your full legal name and your electronic or physical
                signature.
              </li>
              <li>
                Deliver this Notice, with all items completed, to SiaCDN&rsquo;s
                Designated Copyright Agent:
              </li>
            </ol>
            <p>
              Eric Florenzano<br />
              1738 Haight Street, Apt 306<br />
              San Francisco, CA 94117<br />
              +1 612 327-6687<br />
              <a href="mailto:eric@maxint.co">eric@maxint.co</a>
            </p>

            <Header as="h2">Collection of Information</Header>

            <Header as="h3">Information You Provide to Us</Header>
            <p>
              We collect information you provide directly to us. For example, we
              collect information when you create an account, subscribe,
              participate in any interactive features of our services, fill out
              a form, request customer support or otherwise communicate with us.
              The types of information we may collect include your name, email
              address, postal address, credit card information and other contact
              or identifying information you choose to provide.
            </p>

            <Header as="h3">
              Information We Collect Automatically When You Use the Services
            </Header>
            <p>
              When you access or use our services, we automatically collect
              information about you, including:
            </p>
            <ul>
              <li>
                <strong>Log Information</strong>: We log information about your
                use of our services, including the type of browser you use,
                access times, pages viewed, your IP address and the page you
                visited before navigating to our services.
              </li>
              <li>
                <strong>Device Information</strong> We collect information about
                the computer you use to access our services, including the
                hardware model, and operating system and version.
              </li>
              <li>
                <strong>
                  Information Collected by Cookies and Other Tracking
                  Technologies
                </strong>{' '}
                We use various technologies to collect information, and this may
                include sending cookies to your computer. Cookies are small data
                files stored on your hard drive or in your device memory that
                helps us to improve our services and your experience, see which
                areas and features of our services are popular and count visits.
                We may also collect information using web beacons (also known as
                &ldquo;tracking pixels&rdquo;). Web beacons are electronic
                images that may be used in our services or emails and to track
                count visits or understand usage and campaign effectiveness. For
                more information about cookies, and how to disable them, please
                see &ldquo;Your Information Choices&rdquo; below.
              </li>
            </ul>

            <Header as="h3">Information We Collect From Other Sources</Header>
            <p>
              In order to provide you with access to the Service, or to provide
              you with better service in general, we may combine information
              obtained from other sources (for example, a third-party developer
              whose application you have authorized) and combine that with
              information we collect through our services.
            </p>

            <Header as="h3">Use of Information</Header>
            <p>We do not use your information for serving up ads.</p>
            <p>
              We use information about you for various purposes, including to:
            </p>
            <ul>
              <li>Provide, maintain and improve our services;</li>
              <li>
                Provide services you request, process transactions and to send
                you related information, including confirmations and invoices,
              </li>
              <li>
                Send you technical notices, updates, security alerts and support
                and administrative messages;
              </li>
              <li>
                Respond to your comments, questions and requests and provide
                customer service;
              </li>
              <li>
                Communicate with you about SiaCDN-related news and information;
              </li>
              <li>
                Monitor and analyze trends, usage and activities in connection
                with our services; and
              </li>
              <li>Personalize and improve our services.</li>
            </ul>
            <p>
              By accessing and using our services, you consent to the processing
              and transfer of your information in and to the United States and
              other countries.
            </p>

            <Header as="h3">Sharing of Information</Header>
            <p>
              Remember, we don&rsquo;t share your information with advertisers.
            </p>
            <p>We may share personal information about you as follows:</p>
            <ul>
              <li>
                With third party vendors and other service providers who need
                access to your information to carry out work on our behalf
                (e.g., our payment processors who handle the payment portion of
                your subscriptions);
              </li>
              <li>
                If we believe disclosure is reasonably necessary to comply with
                any applicable law, regulation, legal process or governmental
                request;
              </li>
              <li>
                To enforce applicable user agreements or policies, including our{' '}
                <a href="/tos">Terms of Service</a>; and to protect SiaCDN, our
                users or the public from harm or illegal activities;
              </li>
              <li>
                In connection with any merger, sale of SiaCDN assets, financing
                or acquisition of all or a portion of our business to another
                company; and
              </li>
              <li>
                If we notify you through our services (or in our privacy policy)
                that the information you provide will be shared in a particular
                manner and you provide such information.
              </li>
            </ul>
            <p>
              We may also share aggregated or anonymized information that does
              not directly identify you.
            </p>

            <Header as="h3">Third Party Analytics</Header>
            <p>
              We may allow third parties to provide analytics services. These
              third parties may use cookies, web beacons and other technologies
              to collect information about your use of the services and other
              websites, including your IP address, web browser, pages viewed,
              time spent on pages, links clicked and conversion information.
              This information may be used by SiaCDN and third parties to, among
              other things, analyze and track data, determine the popularity of
              certain content and other websites and better understand your
              online activity. Our privacy policy does not apply to, and we are
              not responsible for, third party cookies, web beacons or other
              tracking technologies and we encourage you to check the privacy
              policies of these third parties to learn more about their privacy
              practices.
            </p>

            <Header as="h3">Security</Header>
            <p>
              SiaCDN takes reasonable measures to help protect personal
              information from loss, theft, misuse and unauthorized access,
              disclosure, alteration and destruction.
            </p>

            <Header as="h2">Your Information Choices</Header>

            <Header as="h3">Account Information</Header>
            <p>
              You may update, correct or delete information about you at any
              time by logging into your online account and modifying your
              personal profile. If you wish to delete or suspend your account,
              please email us at{' '}
              <a href="mailto:eric@maxint.co">eric@maxint.co</a>, but note that
              we may retain certain information as required by law or for
              legitimate business purposes. We may also retain cached or
              archived copies of your information for a certain period of time.
              In any case, deleted content will not remain in our system or
              backups for more than two (2) months.
            </p>

            <Header as="h3">Cookies</Header>
            <p>
              Most web browsers are set to accept cookies by default. If you
              prefer, you can usually choose to set your browser to remove or
              reject browser cookies. Please note that if you choose to remove
              or reject cookies, this could affect the availability and
              functionality of our services.
            </p>

            <Header as="h3">Promotional Communications</Header>
            <p>
              You may opt out of receiving promotional emails from SiaCDN by
              following the instructions in those emails. If you opt out, we may
              still send you non-promotional communications, such as those about
              your account or our ongoing business relations.
            </p>

            <Header as="h2">Contact Us</Header>
            <p>
              If you have any questions about this privacy policy, please
              contact us at:
            </p>
            <p>
              SiaCDN Privacy Team<br />
              1738 Haight Street, Apt 306<br />
              San Francisco, CA 94117<br />
              <a href="mailto:eric@maxint.co">eric@maxint.co</a>
            </p>
          </Segment>
          <Footer activeItem="copyright" authAccount={null} />
        </div>
      </div>
    );
  }
}