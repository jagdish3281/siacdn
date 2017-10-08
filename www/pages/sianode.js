import Link from 'next/link';
import Router from 'next/router';
import cookies from 'next-cookies';
import { Segment, Header, Button, List, Card, Icon } from 'semantic-ui-react';
import TimeAgo from 'timeago-react';
import Client from '../lib/client';
import redirect from '../lib/redirect';
import Nav from '../components/nav';
import Footer from '../components/footer';
import { displayStatus } from '../lib/fmt';
import PageHeader from '../components/pageheader';

const instanceArray = siaNode => {
  var resp = [];
  for (var i = 1; i <= siaNode.minio_instances_requested; ++i) {
    resp.push(i);
  }
  return resp;
};

const SiaNode = ({ authTokenID, authAccount, siaNode }) => (
  <div>
    <PageHeader />
    <div className="holder">
      <Nav activeItem="sianode" authAccount={authAccount} />

      <Segment padded>
        <Header as="h2">Sia full node: {siaNode.shortcode}</Header>
        <List>
          <List.Item>
            <List.Content>
              <Icon name="tag" /> <strong>ID</strong>: {siaNode.id}
            </List.Content>
          </List.Item>
          <List.Item>
            <List.Content>
              <Icon name="signal" /> <strong>Status</strong>:{' '}
              {displayStatus[siaNode.status]}
            </List.Content>
          </List.Item>
          <List.Item>
            <List.Content>
              <Icon name="time" /> <strong>Created</strong>:{' '}
              <TimeAgo datetime={siaNode.created_time} />
            </List.Content>
          </List.Item>
          <List.Item>
            <List.Content>
              <Icon name="cloud" /> {siaNode.minio_instances_requested} Minio
              instance{siaNode.minio_instances_requested === 1 ? '' : 's'}
            </List.Content>
          </List.Item>
        </List>
      </Segment>

      <Segment padded>
        <Header as="h2">
          Minio Node{siaNode.minio_instances_requested === 1 ? '' : 's'}
        </Header>
        <List>
          <List.Item>
            <List.Content>
              <Icon name="unlock alternate" /> <strong>Minio Access Key</strong>:{' '}
              {siaNode.minio_access_key}
            </List.Content>
          </List.Item>
          <List.Item>
            <List.Content>
              <Icon name="lock" /> <strong>Minio Secret Key</strong>:{' '}
              {siaNode.minio_secret_key}
            </List.Content>
          </List.Item>
        </List>
        <List>
          {instanceArray(siaNode).map(i => (
            <List.Item key={i}>
              <List.Content>
                <Icon name="linkify" />{' '}
                <a
                  href={`https://${siaNode.shortcode}-minio${i}.siacdn.com`}
                  target="_blank"
                >
                  https://{siaNode.shortcode}-minio{i}.siacdn.com
                </a>
              </List.Content>
            </List.Item>
          ))}
        </List>
      </Segment>

      {siaNode.status === 'stopping' || siaNode.status === 'stopped' ? (
        <Button disabled>Deleting...</Button>
      ) : (
        <Button
          onClick={async ev => {
            if (!confirm('Are you sure you want to delete this Sia node?')) {
              return;
            }
            const client = new Client(authTokenID);
            try {
              await client.deleteSiaNode(siaNode.id);
              Router.push('/dashboard');
            } catch (error) {
              alert(error);
            }
          }}
        >
          Delete SiaNode
        </Button>
      )}

      <Footer activeItem="dashboard" authAccount={authAccount} />
    </div>
  </div>
);

SiaNode.getInitialProps = async ctx => {
  const { authTokenID } = cookies(ctx);
  const { query: { id } } = ctx;
  const client = new Client(authTokenID);
  let authAccount = null;
  try {
    authAccount = await client.getAuthAccount();
    if (!authAccount) {
      redirect(ctx, '/login');
      return { authAccount };
    }
  } catch (err) {
    redirect(ctx, '/login');
    return { authAccount };
  }
  const siaNode = await client.getSiaNode(id);
  return { authTokenID, authAccount, siaNode };
};

export default SiaNode;
